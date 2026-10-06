package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// TraceContextMiddleware extracts the incoming W3C trace context
// (traceparent/tracestate) from the HTTP request headers and attaches it to
// the request context, so spans created by the gateway become children of the
// trace started upstream (e.g. NGINX or the client). When no trace context is
// present a new trace is started by the span created in the route handler.
func TraceContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := otel.GetTextMapPropagator().Extract(
				c.Request().Context(),
				propagation.HeaderCarrier(c.Request().Header),
			)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
