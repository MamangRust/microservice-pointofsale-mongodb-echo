package handler

import (
	"net/http"
	"strconv"
	"strings"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	authpb "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
)

type authHandleApi struct {
	client     authpb.AuthServiceClient
	logger     logger.LoggerInterface
	mapping    response_api.AuthResponseMapper
	apiHandler errors.ApiHandler
	cache      auth_cache.AuthMencache
}

func NewHandlerAuth(
	router *echo.Echo,
	client authpb.AuthServiceClient,
	logger logger.LoggerInterface,
	mapper response_api.AuthResponseMapper,
	apiHandler errors.ApiHandler,
	cache auth_cache.AuthMencache,
) *authHandleApi {
	authHandler := &authHandleApi{
		client:     client,
		logger:     logger,
		mapping:    mapper,
		apiHandler: apiHandler,
		cache:      cache,
	}

	routerAuth := router.Group("/api/auth")

	routerAuth.GET("/verify-code", apiHandler.Handle("get-auth-verifycode", authHandler.VerifyCode))
	routerAuth.POST("/forgot-password", apiHandler.Handle("post-auth-forgotpassword", authHandler.ForgotPassword))
	routerAuth.POST("/reset-password", apiHandler.Handle("post-auth-resetpassword", authHandler.ResetPassword))
	routerAuth.GET("/hello", apiHandler.Handle("get-auth-handlehello", authHandler.HandleHello))
	routerAuth.POST("/register", apiHandler.Handle("post-auth-register", authHandler.Register))
	routerAuth.POST("/login", apiHandler.Handle("post-auth-login", authHandler.Login))
	routerAuth.POST("/refresh-token", apiHandler.Handle("post-auth-refreshtoken", authHandler.RefreshToken))
	routerAuth.GET("/me", apiHandler.Handle("get-auth-getme", authHandler.GetMe))

	return authHandler
}

// Health check
// Health check
// Health check
// @Summary Health check
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/hello [get]
func (h *authHandleApi) HandleHello(c echo.Context) error {
	return c.String(200, "Hello")
}

// Verify reset code
// Verify reset code
// Verify reset code
// @Summary Verify reset code
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseVerifyCode
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/verify-code [get]
func (h *authHandleApi) VerifyCode(c echo.Context) error {
	ctx := c.Request().Context()

	verifyCode, err := parseQueryStringRequired(c, "verify_code")
	if err != nil {
		return errors.NewBadRequestError("invalid verify_code")
	}

	res, err := h.client.VerifyCode(ctx, &authpb.VerifyCodeRequest{
		Code: verifyCode,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToResponseVerifyCode(res)
	return c.JSON(http.StatusOK, so)
}

// Request password reset
// Request password reset
// Request password reset
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.ForgotPasswordRequest true "Request body"
// @Success 200 {object} response.ApiResponseForgotPassword
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/forgot-password [post]
func (h *authHandleApi) ForgotPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var body requests.ForgotPasswordRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}

	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	res, err := h.client.ForgotPassword(ctx, &authpb.ForgotPasswordRequest{
		Email: body.Email,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	resp := h.mapping.ToResponseForgotPassword(res)
	return c.JSON(http.StatusOK, resp)
}

// Reset password
// Reset password
// Reset password
// @Summary Reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.CreateResetPasswordRequest true "Request body"
// @Success 200 {object} response.ApiResponseResetPassword
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/reset-password [post]
func (h *authHandleApi) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var body requests.CreateResetPasswordRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}

	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	res, err := h.client.ResetPassword(ctx, &authpb.ResetPasswordRequest{
		ResetToken:      body.ResetToken,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToResponseResetPassword(res)
	return c.JSON(http.StatusOK, so)
}

// Register a new user
// Register a new user
// Register a new user
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "Request body"
// @Success 200 {object} response.ApiResponseRegister
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/register [post]
func (h *authHandleApi) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var body requests.CreateUserRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}

	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	data := &authpb.RegisterRequest{
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.client.RegisterUser(ctx, data)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToResponseRegister(res)
	return c.JSON(http.StatusOK, so)
}

// Login
// Login
// Login
// @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.AuthRequest true "Request body"
// @Success 200 {object} response.ApiResponseLogin
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/login [post]
func (h *authHandleApi) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var body requests.AuthRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}

	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	// Login Cache Check
	if cached, found := h.cache.GetCachedLogin(ctx, body.Email); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	data := &authpb.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	}

	res, err := h.client.LoginUser(ctx, data)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mappedResponse := h.mapping.ToResponseLogin(res)

	// Set Login Cache
	h.cache.SetCachedLogin(ctx, body.Email, mappedResponse)

	return c.JSON(http.StatusOK, mappedResponse)
}

// Refresh access token
// Refresh access token
// Refresh access token
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenRequest true "Request body"
// @Success 200 {object} response.ApiResponseRefreshToken
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/refresh-token [post]
func (h *authHandleApi) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	var body requests.RefreshTokenRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}

	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	// Refresh Token Cache Check
	if cached, found := h.cache.GetRefreshToken(ctx, body.RefreshToken); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.client.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToResponseRefreshToken(res)

	// Set Refresh Token Cache
	h.cache.SetRefreshToken(ctx, body.RefreshToken, so)

	return c.JSON(http.StatusOK, so)
}

// Get current user profile
// Get current user profile
// Get current user profile
// @Summary Get current user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseGetMe
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/me [get]
// @Security BearerAuth
func (h *authHandleApi) GetMe(c echo.Context) error {
	ctx := c.Request().Context()

	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.ErrUnauthorized.WithMessage("invalid authorization header")
	}

	// Extract userID from context if set by auth middleware
	var userID string
	if uVal := c.Get("userID"); uVal != nil {
		if uStr, ok := uVal.(string); ok {
			userID = uStr
		}
	}

	// User Info Cache Check
	if userID != "" {
		if cached, found := h.cache.GetCachedUserInfo(ctx, userID); found && cached != nil {
			return c.JSON(http.StatusOK, cached)
		}
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	res, err := h.client.GetMe(ctx, &authpb.GetMeRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToResponseGetMe(res)

	// Set User Info Cache
	if userID == "" && so.Data != nil {
		userID = strconv.Itoa(so.Data.ID)
	}
	if userID != "" {
		h.cache.SetCachedUserInfo(ctx, userID, so)
	}

	return c.JSON(http.StatusOK, so)
}

func parseQueryStringRequired(c echo.Context, name string) (string, error) {
	val := c.QueryParam(name)
	if val == "" {
		return "", errors.NewBadRequestError("missing " + name)
	}
	return val, nil
}
