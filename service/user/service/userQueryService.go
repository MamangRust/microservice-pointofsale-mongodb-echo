package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	mencache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	"github.com/MamangRust/microservice-point-of-sale-user/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type userQueryDeps struct {
	Cache         mencache.UserQueryCache
	UserQuery     repository.UserQueryRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type userQueryService struct {
	mencache      mencache.UserQueryCache
	userQuery     repository.UserQueryRepository
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewUserQueryService(params *userQueryDeps) UserQueryService {
	return &userQueryService{
		mencache:      params.Cache,
		userQuery:     params.UserQuery,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *userQueryService) FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, error) {
	const method = "FindAll"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", req.Search))
	defer func() { end(status) }()

	if data, total, found := s.mencache.GetCachedUsersCache(ctx, req); found {
		logSuccess("Successfully fetched users from cache", zap.Int("totalRecords", *total), zap.Int("page", page))
		return data, total, nil
	}

	users, err := s.userQuery.FindAllUsers(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*repository.UserResult](
			s.logger, user_errors.ErrFailedFindAll.WithInternal(err), method, span, zap.Error(err))
	}

	var totalCount int
	if len(users) > 0 {
		totalCount = int(users[0].TotalCount)
	}
	s.mencache.SetCachedUsersCache(ctx, req, users, &totalCount)
	logSuccess("Successfully fetched users", zap.Int("totalRecords", totalCount), zap.Int("page", page))
	return users, &totalCount, nil
}

func (s *userQueryService) FindByID(ctx context.Context, id int) (*models.User, error) {
	const method = "FindByID"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("user.id", id))
	defer func() { end(status) }()

	if data, found := s.mencache.GetCachedUserCache(ctx, id); found {
		logSuccess("Successfully fetched user from cache", zap.Int("user.id", id))
		return data, nil
	}

	user, err := s.userQuery.FindById(ctx, id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span, zap.Int("user.id", id))
	}

	s.mencache.SetCachedUserCache(ctx, user)
	logSuccess("Successfully fetched user", zap.Int("user.id", id))
	return user, nil
}

func (s *userQueryService) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, error) {
	const method = "FindByActive"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", req.Search))
	defer func() { end(status) }()

	if data, total, found := s.mencache.GetCachedUserActiveCache(ctx, req); found {
		logSuccess("Successfully fetched active users from cache", zap.Int("totalRecords", *total), zap.Int("page", page))
		return data, total, nil
	}

	users, err := s.userQuery.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*repository.UserResult](
			s.logger, user_errors.ErrFailedFindActive.WithInternal(err), method, span, zap.Error(err))
	}

	var totalCount int
	if len(users) > 0 {
		totalCount = int(users[0].TotalCount)
	}
	s.mencache.SetCachedUserActiveCache(ctx, req, users, &totalCount)
	logSuccess("Successfully fetched active users", zap.Int("totalRecords", totalCount), zap.Int("page", page))
	return users, &totalCount, nil
}

func (s *userQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, error) {
	const method = "FindByTrashed"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", req.Search))
	defer func() { end(status) }()

	if data, total, found := s.mencache.GetCachedUserTrashedCache(ctx, req); found {
		logSuccess("Successfully fetched trashed users from cache", zap.Int("totalRecords", *total), zap.Int("page", page))
		return data, total, nil
	}

	users, err := s.userQuery.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*repository.UserResult](
			s.logger, user_errors.ErrFailedFindTrashed.WithInternal(err), method, span, zap.Error(err))
	}

	var totalCount int
	if len(users) > 0 {
		totalCount = int(users[0].TotalCount)
	}
	s.mencache.SetCachedUserTrashedCache(ctx, req, users, &totalCount)
	logSuccess("Successfully fetched trashed users", zap.Int("totalRecords", totalCount), zap.Int("page", page))
	return users, &totalCount, nil
}

func (s *userQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

// FindByEmailWithPassword serves the auth service's credential lookups. The
// password hash is deliberately never cached.
func (s *userQueryService) FindByEmailWithPassword(ctx context.Context, email string) (*models.User, error) {
	const method = "FindByEmailWithPassword"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("email", email))
	defer func() { end(status) }()

	user, err := s.userQuery.FindByEmailWithPassword(ctx, email)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span, zap.String("email", email))
	}
	logSuccess("Successfully fetched user with password", zap.String("email", email))
	return user, nil
}

func (s *userQueryService) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	const method = "FindByEmailAndVerify"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("email", email))
	defer func() { end(status) }()

	user, err := s.userQuery.FindByEmailAndVerify(ctx, email)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span, zap.String("email", email))
	}
	logSuccess("Successfully fetched verified user", zap.String("email", email))
	return user, nil
}

func (s *userQueryService) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	const method = "FindByVerificationCode"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Bool("verification_code_present", code != ""))
	defer func() { end(status) }()

	user, err := s.userQuery.FindByVerificationCode(ctx, code)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, user_errors.ErrFailedFindUserByCode.WithInternal(err), method, span)
	}
	logSuccess("Successfully fetched user by verification code")
	return user, nil
}
