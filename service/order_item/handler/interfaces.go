package handler

import (
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
)

type OrderItemQueryHandleGrpc interface {
	pborderitem.OrderItemServiceServer
}

type OrderItemCommandHandleGrpc interface {
	pborderitem.OrderItemServiceServer
}
