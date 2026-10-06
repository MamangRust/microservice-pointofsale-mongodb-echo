package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type orderCommandDeps struct {
	Cache                      mencache.OrderCommandCache
	CashierQueryRepository     cashieradapter.QueryRepository
	OrderQueryRepository       repository.OrderQueryRepository
	OrderCommandRepository     repository.OrderCommandRepository
	OrderItemQueryRepository   orderitemadapter.QueryRepository
	OrderItemCommandRepository orderitemadapter.CommandRepository
	MerchantQueryRepository    merchantadapter.Repository
	ProductQueryRepository     productadapter.QueryRepository
	ProductCommandRepository   productadapter.CommandRepository
	Outbox                     *outbox.OutboxService
	Logger                     logger.LoggerInterface
	Observability              observability.TraceLoggerObservability
}

type stockChange struct {
	productID int
	quantity  int
}

type stockMutation struct {
	productID int
	quantity  int
	increment bool
}

type orderCommandService struct {
	mencache                   mencache.OrderCommandCache
	cashierQueryRepository     cashieradapter.QueryRepository
	orderQueryRepository       repository.OrderQueryRepository
	orderCommandRepository     repository.OrderCommandRepository
	orderItemQueryRepository   orderitemadapter.QueryRepository
	orderItemCommandRepository orderitemadapter.CommandRepository
	merchantQueryRepository    merchantadapter.Repository
	productQueryRepository     productadapter.QueryRepository
	productCommandRepository   productadapter.CommandRepository
	outbox                     *outbox.OutboxService
	logger                     logger.LoggerInterface
	observability              observability.TraceLoggerObservability
}

func NewOrderCommandService(params *orderCommandDeps) OrderCommandService {
	return &orderCommandService{
		mencache:                   params.Cache,
		cashierQueryRepository:     params.CashierQueryRepository,
		orderQueryRepository:       params.OrderQueryRepository,
		orderCommandRepository:     params.OrderCommandRepository,
		orderItemQueryRepository:   params.OrderItemQueryRepository,
		orderItemCommandRepository: params.OrderItemCommandRepository,
		merchantQueryRepository:    params.MerchantQueryRepository,
		productQueryRepository:     params.ProductQueryRepository,
		productCommandRepository:   params.ProductCommandRepository,
		outbox:                     params.Outbox,
		logger:                     params.Logger,
		observability:              params.Observability,
	}
}

func (s *orderCommandService) CreateOrder(ctx context.Context, req *requests.CreateOrderRequest) (*models.Order, error) {
	const method = "CreateOrder"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("merchant.id", req.MerchantID),
		attribute.Int("cashier.id", req.CashierID),
	)
	defer func() {
		end(status)
	}()

	_, err := s.merchantQueryRepository.FindById(ctx, req.MerchantID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](
			s.logger,
			merchant_errors.ErrFailedFindMerchantById.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	_, err = s.cashierQueryRepository.FindById(ctx, req.CashierID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](
			s.logger,
			cashier_errors.ErrFailedFindCashierById.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	products := make([]*models.Product, 0, len(req.Items))
	for _, item := range req.Items {
		product, err := s.productQueryRepository.FindById(ctx, item.ProductID)
		if err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, product_errors.ErrFailedFindProductById.WithInternal(err), method, span, zap.Error(err))
		}
		if item.Quantity <= 0 {
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("quantity must be positive"), method, span, zap.Int("product.id", item.ProductID))
		}
		products = append(products, product)
	}

	order, err := s.orderCommandRepository.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: req.MerchantID,
		CashierID:  req.CashierID,
	})
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedCreateOrder.WithInternal(err), method, span, zap.Error(err))
	}

	span.SetAttributes(attribute.Int("order.id", int(order.OrderID)))
	applied := make([]stockChange, 0, len(req.Items))
	createdItems := make([]int, 0, len(req.Items))
	compensate := func() {
		for i := len(createdItems) - 1; i >= 0; i-- {
			if rollbackErr := s.orderItemCommandRepository.DeleteOrderItem(ctx, createdItems[i]); rollbackErr != nil {
				s.logger.Error("failed to delete compensated order item", zap.Int("order_item.id", createdItems[i]), zap.Error(rollbackErr))
			}
		}
		if rollbackErr := s.orderCommandRepository.DeleteOrder(ctx, int(order.OrderID)); rollbackErr != nil {
			s.logger.Error("failed to delete compensated order", zap.Int("order.id", int(order.OrderID)), zap.Error(rollbackErr))
		}
		for i := len(applied) - 1; i >= 0; i-- {
			if _, rollbackErr := s.productCommandRepository.IncrementProductCountStock(ctx, applied[i].productID, applied[i].quantity); rollbackErr != nil {
				s.logger.Error("failed to compensate order stock", zap.Int("product.id", applied[i].productID), zap.Int("quantity", applied[i].quantity), zap.Error(rollbackErr))
			}
		}
	}
	for i, item := range req.Items {
		if _, err = s.productCommandRepository.DecrementProductCountStock(ctx, int(products[i].ProductID), item.Quantity); err != nil {
			compensate()
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, err, method, span, zap.Int("product.id", item.ProductID), zap.Error(err))
		}
		applied = append(applied, stockChange{productID: int(products[i].ProductID), quantity: item.Quantity})
	}

	for i, item := range req.Items {
		createdItem, createErr := s.orderItemCommandRepository.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{OrderID: int(order.OrderID), ProductID: item.ProductID, Quantity: item.Quantity, Price: int(products[i].Price)})
		if createErr != nil {
			compensate()
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedCreateOrderItem.WithInternal(createErr), method, span, zap.Error(createErr))
		}
		if createdItem == nil {
			compensate()
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("order item repository returned nil item"), method, span)
		}
		createdItems = append(createdItems, int(createdItem.OrderItemID))

		// F3 — stats pipeline: publish a stats.pos.order_item.event so
		// stats-writer materializes this line item into ClickHouse for the
		// product/category sold statistics. CategoryID comes from the product
		// (denormalized here so the reader never does a cross-collection lookup). Best-effort:
		// the order flow spans services via gRPC, so there is no single DB tx
		// that could carry the event atomically; failures are logged and can be
		// reconciled via backfill.
		s.publishStatsEvent(ctx, "stats.pos.order_item.event", strconv.Itoa(int(createdItem.OrderItemID)), events.OrderItemEvent{
			OrderItemID: createdItem.OrderItemID,
			OrderID:     order.OrderID,
			ProductID:   createdItem.ProductID,
			CategoryID:  products[i].CategoryID,
			Quantity:    createdItem.Quantity,
			UnitPrice:   int32(createdItem.Price),
			Subtotal:    createdItem.Quantity * int32(createdItem.Price),
			EventTime:   time.Now().UTC().Format(time.RFC3339),
		})
	}

	totalPrice, err := s.orderItemQueryRepository.CalculateTotalPrice(ctx, int(order.OrderID))
	if err != nil {
		compensate()
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](
			s.logger,
			orderitem_errors.ErrFailedCalculateTotal.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	res, err := s.orderCommandRepository.UpdateOrder(ctx, &requests.UpdateOrderRecordRequest{
		OrderID:    int(order.OrderID),
		TotalPrice: int(*totalPrice),
	})
	if err != nil {
		compensate()
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](
			s.logger,
			order_errors.ErrFailedUpdateOrder.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	logSuccess("Successfully created order", zap.Int("order.id", int(res.OrderID)))

	// F3 — stats pipeline: publish a stats.pos.order.event so stats-writer
	// materializes this order into ClickHouse for the order revenue statistics.
	// Status mirrors the backfill convention ("created").
	s.publishStatsEvent(ctx, "stats.pos.order.event", strconv.Itoa(int(res.OrderID)), events.OrderEvent{
		OrderID:    res.OrderID,
		CashierID:  res.CashierID,
		MerchantID: res.MerchantID,
		TotalPrice: res.TotalPrice,
		Status:     "created",
		EventTime:  time.Now().UTC().Format(time.RFC3339),
	})

	return res, nil
}

// publishStatsEvent marshals a stats domain event into the standard envelope
// (event_id for idempotency + business payload) and enqueues it in the outbox
// for the relay to publish to Kafka. It is best-effort: the order flow spans
// multiple services via gRPC, so there is no single database transaction that
// could carry the event atomically; a failure here is logged and must be
// reconciled manually or via backfill, never by rolling back the order.
func (s *orderCommandService) publishStatsEvent(ctx context.Context, topic, key string, payload any) {
	if s.outbox == nil {
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal stats event payload", zap.Error(err), zap.String("topic", topic))
		return
	}
	envelope, err := json.Marshal(events.StatsEnvelope{
		EventID: uuid.NewString(),
		Payload: payloadJSON,
	})
	if err != nil {
		s.logger.Error("failed to marshal stats event envelope", zap.Error(err), zap.String("topic", topic))
		return
	}
	if err := s.outbox.Enqueue(ctx, topic, key, envelope); err != nil {
		s.logger.Error("failed to enqueue stats event", zap.Error(err), zap.String("topic", topic))
	}
}

func (s *orderCommandService) UpdateOrder(ctx context.Context, req *requests.UpdateOrderRequest) (*models.Order, error) {
	const method = "UpdateOrder"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("order.id", *req.OrderID))
	defer func() { end(status) }()

	if _, err := s.orderQueryRepository.FindById(ctx, *req.OrderID); err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedFindOrderById.WithInternal(err), method, span, zap.Error(err))
	}

	oldItems, err := s.orderItemQueryRepository.FindOrderItemByOrder(ctx, *req.OrderID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedFindOrderItemByOrder.WithInternal(err), method, span, zap.Error(err))
	}
	oldByID := make(map[int32]*models.OrderItem, len(oldItems))
	for _, item := range oldItems {
		oldByID[item.OrderItemID] = item
	}
	seenItemIDs := make(map[int]struct{}, len(req.Items))

	products := make([]*models.Product, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("quantity must be positive"), method, span)
		}
		product, findErr := s.productQueryRepository.FindById(ctx, item.ProductID)
		if findErr != nil {
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, product_errors.ErrFailedFindProductById.WithInternal(findErr), method, span, zap.Error(findErr))
		}
		products = append(products, product)
	}

	changes := make([]stockChange, 0, len(req.Items)*2)
	updatedItems := make([]*models.OrderItem, 0, len(req.Items))
	createdItemIDs := make([]int, 0, len(req.Items))
	rollbackEntities := func() {
		for i := len(createdItemIDs) - 1; i >= 0; i-- {
			if rollbackErr := s.orderItemCommandRepository.DeleteOrderItem(ctx, createdItemIDs[i]); rollbackErr != nil {
				s.logger.Error("failed to rollback created order item", zap.Int("order_item.id", createdItemIDs[i]), zap.Error(rollbackErr))
			}
		}
		for i := len(updatedItems) - 1; i >= 0; i-- {
			item := updatedItems[i]
			if _, rollbackErr := s.orderItemCommandRepository.UpdateOrderItem(ctx, &requests.UpdateOrderItemRecordRequest{OrderItemID: int(item.OrderItemID), OrderID: int(item.OrderID), ProductID: int(item.ProductID), Quantity: int(item.Quantity), Price: int(item.Price)}); rollbackErr != nil {
				s.logger.Error("failed to rollback updated order item", zap.Int("order_item.id", int(item.OrderItemID)), zap.Error(rollbackErr))
			}
		}
	}
	rollback := func() {
		rollbackEntities()
		for i := len(changes) - 1; i >= 0; i-- {
			change := changes[i]
			var rollbackErr error
			if change.quantity > 0 {
				_, rollbackErr = s.productCommandRepository.IncrementProductCountStock(ctx, change.productID, change.quantity)
			} else {
				_, rollbackErr = s.productCommandRepository.DecrementProductCountStock(ctx, change.productID, -change.quantity)
			}
			if rollbackErr != nil {
				s.logger.Error("failed to rollback order stock change", zap.Int("product.id", change.productID), zap.Int("quantity", change.quantity), zap.Error(rollbackErr))
			}
		}
	}
	apply := func(productID, quantity int, decrement bool) error {
		if quantity <= 0 {
			return nil
		}
		if decrement {
			if _, applyErr := s.productCommandRepository.DecrementProductCountStock(ctx, productID, quantity); applyErr != nil {
				return applyErr
			}
		} else if _, applyErr := s.productCommandRepository.IncrementProductCountStock(ctx, productID, quantity); applyErr != nil {
			return applyErr
		}
		changes = append(changes, stockChange{productID: productID, quantity: func() int {
			if decrement {
				return quantity
			}
			return -quantity
		}()})
		return nil
	}

	for i, item := range req.Items {
		if item.OrderItemID > 0 {
			if _, duplicate := seenItemIDs[item.OrderItemID]; duplicate {
				status = "error"
				return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("order item %d appears more than once", item.OrderItemID), method, span)
			}
			seenItemIDs[item.OrderItemID] = struct{}{}
		}
		old := oldByID[int32(item.OrderItemID)]
		if item.OrderItemID > 0 && old == nil {
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("order item %d not found", item.OrderItemID), method, span)
		}
		if old != nil {
			if old.ProductID == int32(item.ProductID) {
				delta := item.Quantity - int(old.Quantity)
				if delta > 0 {
					err = apply(item.ProductID, delta, true)
				} else {
					err = apply(item.ProductID, -delta, false)
				}
			} else {
				err = apply(int(old.ProductID), int(old.Quantity), false)
				if err == nil {
					err = apply(item.ProductID, item.Quantity, true)
				}
			}
		} else {
			err = apply(item.ProductID, item.Quantity, true)
		}
		if err != nil {
			rollback()
			status = "error"
			return sharederrorhandler.HandleError[*models.Order](s.logger, err, method, span, zap.Error(err))
		}
		if old != nil {
			if _, err = s.orderItemCommandRepository.UpdateOrderItem(ctx, &requests.UpdateOrderItemRecordRequest{OrderItemID: item.OrderItemID, OrderID: *req.OrderID, ProductID: item.ProductID, Quantity: item.Quantity, Price: int(products[i].Price)}); err != nil {
				rollback()
				status = "error"
				return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedUpdateOrderItem.WithInternal(err), method, span, zap.Error(err))
			}
			updatedItems = append(updatedItems, old)
		} else {
			createdItem, createErr := s.orderItemCommandRepository.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{OrderID: *req.OrderID, ProductID: item.ProductID, Quantity: item.Quantity, Price: int(products[i].Price)})
			if createErr != nil {
				rollback()
				status = "error"
				return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedCreateOrderItem.WithInternal(createErr), method, span, zap.Error(createErr))
			}
			if createdItem == nil {
				rollback()
				status = "error"
				return sharederrorhandler.HandleError[*models.Order](s.logger, fmt.Errorf("order item repository returned nil item"), method, span)
			}
			createdItemIDs = append(createdItemIDs, int(createdItem.OrderItemID))

			// F3 — stats pipeline: newly added line items on an existing order
			// feed the product/category sold statistics the same way as on
			// create. Best-effort via the outbox.
			s.publishStatsEvent(ctx, "stats.pos.order_item.event", strconv.Itoa(int(createdItem.OrderItemID)), events.OrderItemEvent{
				OrderItemID: createdItem.OrderItemID,
				OrderID:     createdItem.OrderID,
				ProductID:   createdItem.ProductID,
				CategoryID:  products[i].CategoryID,
				Quantity:    createdItem.Quantity,
				UnitPrice:   int32(createdItem.Price),
				Subtotal:    createdItem.Quantity * int32(createdItem.Price),
				EventTime:   time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	totalPrice, err := s.orderItemQueryRepository.CalculateTotalPrice(ctx, *req.OrderID)
	if err != nil {
		rollback()
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedCalculateTotal.WithInternal(err), method, span, zap.Error(err))
	}
	res, err := s.orderCommandRepository.UpdateOrder(ctx, &requests.UpdateOrderRecordRequest{OrderID: *req.OrderID, TotalPrice: int(*totalPrice)})
	if err != nil {
		rollback()
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedUpdateOrder.WithInternal(err), method, span, zap.Error(err))
	}

	s.mencache.DeleteOrderCache(ctx, *req.OrderID)
	logSuccess("Successfully updated order", zap.Int("order.id", *req.OrderID))
	return res, nil
}

func (s *orderCommandService) reverseStockMutations(ctx context.Context, applied []stockMutation) {
	for i := len(applied) - 1; i >= 0; i-- {
		mutation := applied[i]
		var err error
		if mutation.increment {
			_, err = s.productCommandRepository.DecrementProductCountStock(ctx, mutation.productID, mutation.quantity)
		} else {
			_, err = s.productCommandRepository.IncrementProductCountStock(ctx, mutation.productID, mutation.quantity)
		}
		if err != nil {
			s.logger.Error("failed to compensate order lifecycle stock", zap.Int("product.id", mutation.productID), zap.Int("quantity", mutation.quantity), zap.Bool("increment", mutation.increment), zap.Error(err))
		}
	}
}

func (s *orderCommandService) incrementStockForOrderItems(ctx context.Context, items []*models.OrderItem) ([]stockMutation, error) {
	applied := make([]stockMutation, 0, len(items))
	for _, item := range items {
		if item == nil || item.Quantity <= 0 {
			continue
		}
		if _, err := s.productCommandRepository.IncrementProductCountStock(ctx, int(item.ProductID), int(item.Quantity)); err != nil {
			s.reverseStockMutations(ctx, applied)
			return nil, err
		}
		applied = append(applied, stockMutation{productID: int(item.ProductID), quantity: int(item.Quantity), increment: true})
	}
	return applied, nil
}

func (s *orderCommandService) decrementStockForOrderItems(ctx context.Context, items []*models.OrderItem) ([]stockMutation, error) {
	applied := make([]stockMutation, 0, len(items))
	for _, item := range items {
		if item == nil || item.Quantity <= 0 {
			continue
		}
		if _, err := s.productCommandRepository.DecrementProductCountStock(ctx, int(item.ProductID), int(item.Quantity)); err != nil {
			s.reverseStockMutations(ctx, applied)
			return nil, err
		}
		applied = append(applied, stockMutation{productID: int(item.ProductID), quantity: int(item.Quantity)})
	}
	return applied, nil
}

func (s *orderCommandService) TrashedOrder(ctx context.Context, orderID int) (*models.Order, error) {
	const method = "TrashedOrder"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("order.id", orderID))
	defer func() { end(status) }()

	order, err := s.orderQueryRepository.FindById(ctx, orderID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedFindOrderById.WithInternal(err), method, span, zap.Error(err))
	}
	if order == nil || order.DeletedAt != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedNotDeleteAtOrder, method, span)
	}

	items, err := s.orderItemQueryRepository.FindOrderItemByOrder(ctx, orderID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedFindOrderItemByOrder.WithInternal(err), method, span, zap.Error(err))
	}
	applied, err := s.incrementStockForOrderItems(ctx, items)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, err, method, span, zap.Error(err))
	}

	trashedOrder, err := s.orderCommandRepository.TrashedOrder(ctx, orderID)
	if err != nil || trashedOrder == nil {
		s.reverseStockMutations(ctx, applied)
		status = "error"
		if err == nil {
			err = order_errors.ErrFailedOrderNotActive
		}
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedTrashOrder.WithInternal(err), method, span, zap.Error(err))
	}

	s.mencache.DeleteOrderCache(ctx, orderID)
	logSuccess("Successfully trashed order", zap.Int("order.id", orderID))
	return trashedOrder, nil
}

func (s *orderCommandService) RestoreOrder(ctx context.Context, orderID int) (*models.Order, error) {
	const method = "RestoreOrder"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("order.id", orderID))
	defer func() { end(status) }()

	trashedOrder, err := s.orderQueryRepository.FindByTrashedId(ctx, orderID)
	if err != nil || trashedOrder == nil || trashedOrder.DeletedAt == nil {
		status = "error"
		if err == nil {
			err = order_errors.ErrFailedOrderNotTrashed
		}
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedOrderNotTrashed.WithInternal(err), method, span, zap.Error(err))
	}

	items, err := s.orderItemQueryRepository.FindOrderItemByOrder(ctx, orderID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, orderitem_errors.ErrFailedFindOrderItemByOrder.WithInternal(err), method, span, zap.Error(err))
	}
	applied, err := s.decrementStockForOrderItems(ctx, items)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.Order](s.logger, err, method, span, zap.Error(err))
	}

	order, err := s.orderCommandRepository.RestoreOrder(ctx, orderID)
	if err != nil || order == nil {
		s.reverseStockMutations(ctx, applied)
		status = "error"
		if err == nil {
			err = order_errors.ErrFailedOrderNotActive
		}
		return sharederrorhandler.HandleError[*models.Order](s.logger, order_errors.ErrFailedRestoreOrder.WithInternal(err), method, span, zap.Error(err))
	}

	s.mencache.DeleteOrderCache(ctx, orderID)
	logSuccess("Successfully restored order", zap.Int("order.id", orderID))
	return order, nil
}

func (s *orderCommandService) DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error) {
	const method = "DeleteOrderPermanent"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("order.id", orderID))
	defer func() { end(status) }()

	order, err := s.orderQueryRepository.FindByTrashedId(ctx, orderID)
	if err != nil || order == nil || order.DeletedAt == nil {
		status = "error"
		if err == nil {
			err = order_errors.ErrFailedOrderNotTrashed
		}
		return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedOrderNotTrashed.WithInternal(err), method, span, zap.Error(err))
	}

	success, err := s.orderCommandRepository.DeleteOrderPermanent(ctx, orderID)
	if err != nil || !success {
		status = "error"
		if err == nil {
			err = order_errors.ErrFailedDeleteOrderPermanent
		}
		return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedDeleteOrderPermanent.WithInternal(err), method, span, zap.Error(err))
	}

	s.mencache.DeleteOrderCache(ctx, orderID)
	logSuccess("Successfully deleted order permanently", zap.Int("order.id", orderID))
	return success, nil
}

func (s *orderCommandService) RestoreAllOrder(ctx context.Context) (bool, error) {
	const method = "RestoreAllOrder"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status) }()

	orders, err := s.orderCommandRepository.FindAllTrashed(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedRestoreAllOrder.WithInternal(err), method, span, zap.Error(err))
	}

	type restoredOrder struct {
		orderID int
		items   []*models.OrderItem
		stock   []stockMutation
	}
	restoredOrders := make([]restoredOrder, 0, len(orders))
	rollbackAll := func() {
		for i := len(restoredOrders) - 1; i >= 0; i-- {
			candidate := restoredOrders[i]
			if _, rollbackErr := s.orderCommandRepository.TrashedOrder(ctx, candidate.orderID); rollbackErr != nil {
				s.logger.Error("failed to rollback restored order state", zap.Int("order.id", candidate.orderID), zap.Error(rollbackErr))
			}
			s.reverseStockMutations(ctx, candidate.stock)
		}
	}

	restoredCount := 0
	for _, candidate := range orders {
		if candidate == nil || candidate.DeletedAt == nil {
			continue
		}
		items, findErr := s.orderItemQueryRepository.FindOrderItemByOrder(ctx, int(candidate.OrderID))
		if findErr != nil {
			rollbackAll()
			status = "error"
			return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedRestoreAllOrder.WithInternal(findErr), method, span, zap.Error(findErr))
		}
		applied, decrementErr := s.decrementStockForOrderItems(ctx, items)
		if decrementErr != nil {
			rollbackAll()
			status = "error"
			return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedRestoreAllOrder.WithInternal(decrementErr), method, span, zap.Error(decrementErr))
		}
		restored, restoreErr := s.orderCommandRepository.RestoreOrder(ctx, int(candidate.OrderID))
		if restoreErr != nil || restored == nil {
			s.reverseStockMutations(ctx, applied)
			if restoreErr != nil && errors.Is(restoreErr, order_errors.ErrRestoreOrderNotFound) {
				continue
			}
			rollbackAll()
			if restoreErr == nil {
				restoreErr = order_errors.ErrFailedOrderNotActive
			}
			status = "error"
			return sharederrorhandler.HandleError[bool](s.logger, order_errors.ErrFailedRestoreAllOrder.WithInternal(restoreErr), method, span, zap.Error(restoreErr))
		}
		restoredOrders = append(restoredOrders, restoredOrder{orderID: int(candidate.OrderID), items: items, stock: applied})
		restoredCount++
	}

	s.mencache.DeleteOrderAllCache(ctx)
	logSuccess("Successfully restored all orders", zap.Int("restored.count", restoredCount))
	return true, nil
}

func (s *orderCommandService) DeleteAllOrderPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllOrderPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.orderCommandRepository.DeleteAllOrderPermanent(ctx)
	if err != nil || !success {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			order_errors.ErrFailedDeleteAllOrderPermanent.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteOrderAllCache(ctx)
	logSuccess("Successfully deleted all orders permanently", zap.Bool("success", success))
	return success, nil
}
