package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/realtime"
	"github.com/mrkaak/restaurant-api/internal/services"
)

// StreamHandler streams live order-status updates over Server-Sent Events
// (plan §9). SSE is simpler than WebSockets and sufficient for one-way status
// pushes.
type StreamHandler struct {
	orders *services.OrderService
	hub    *realtime.Hub
}

func NewStreamHandler(o *services.OrderService, hub *realtime.Hub) *StreamHandler {
	return &StreamHandler{orders: o, hub: hub}
}

func (h *StreamHandler) Register(api *echo.Group, jwtAuth echo.MiddlewareFunc) {
	api.GET("/orders/:code/stream", h.Stream, jwtAuth)
}

// Stream emits the current status immediately, then one SSE message per status
// change until the client disconnects.
func (h *StreamHandler) Stream(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	role, _ := middleware.RoleFrom(c)
	code := c.Param("code")

	// Authorize: the order must belong to the caller (or the caller is staff).
	order, err := h.orders.GetByCode(c.Request().Context(), uid, role.IsStaff(), code)
	if err != nil {
		return mapServiceError(c, err)
	}

	w := c.Response()
	w.Header().Set(echo.HeaderContentType, "text/event-stream")
	w.Header().Set(echo.HeaderCacheControl, "no-cache")
	w.Header().Set(echo.HeaderConnection, "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.Writer.(http.Flusher)
	if !ok {
		return c.String(http.StatusInternalServerError, "streaming unsupported")
	}

	events, unsubscribe := h.hub.Subscribe(code)
	defer unsubscribe()

	// Send the current status right away so a late subscriber is in sync.
	writeSSE(w, flusher, "status", fmt.Sprintf(`{"order_code":%q,"status":%q}`, order.Code, order.Status))

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-events:
			if !ok {
				return nil
			}
			writeSSE(w, flusher, "status", fmt.Sprintf(`{"order_code":%q,"status":%q,"message":%q}`, e.OrderCode, e.Status, e.Message))
		}
	}
}

func writeSSE(w *echo.Response, f http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	f.Flush()
}
