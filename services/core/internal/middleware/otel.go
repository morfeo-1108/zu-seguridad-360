package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"time"
)

var tracer = otel.Tracer("zu-core")
var logger, _ = zap.NewProduction()

func OtelTracing() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), c.Path())
		defer span.End()

		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
		}
		if tenantID, ok := c.Locals("tenant_id").(string); ok && tenantID != "" {
			attrs = append(attrs, attribute.String("tenant.id", tenantID))
		}
		span.SetAttributes(attrs...)

		c.SetUserContext(ctx)
		return c.Next()
	}
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)
		logger.Info("request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", requestID),
		)

		return err
	}
}
