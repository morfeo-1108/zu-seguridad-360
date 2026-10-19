package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/service"
)

type EventLogHandler struct{ svc *service.EventLogService }

func NewEventLogHandler(svc *service.EventLogService) *EventLogHandler {
	return &EventLogHandler{svc: svc}
}

// Append — registra un evento de agente (lo llaman los servicios de agentes).
func (h *EventLogHandler) Append(c *fiber.Ctx) error {
	var ev model.AgentEvent
	if err := c.BodyParser(&ev); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	ev.TenantID = mustUUID(c.Locals("tenant_id"))
	if err := h.svc.Append(c.Context(), &ev); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(ev)
}

// Recent — el REGISTRO (bitácora) de eventos, filtrable por proyecto.
func (h *EventLogHandler) Recent(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	var projectID *uuid.UUID
	if p := c.Query("project_id"); p != "" {
		if id, err := uuid.Parse(p); err == nil {
			projectID = &id
		}
	}
	events, err := h.svc.Recent(c.Context(), tenantID, projectID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": events, "total": len(events)})
}

// Stream — feed EN VIVO vía Server-Sent Events (sin polling).
// El frontend abre un EventSource a este endpoint y recibe eventos push.
func (h *EventLogHandler) Stream(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id")
	tid := ""
	if tenantID != nil {
		tid = fmt.Sprintf("%v", tenantID)
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ch := h.svc.Subscribe(tid)

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer h.svc.Unsubscribe(tid, ch)
		// heartbeat para mantener viva la conexión
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				data, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: agent_event\ndata: %s\n\n", data)
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	}))
	return nil
}
