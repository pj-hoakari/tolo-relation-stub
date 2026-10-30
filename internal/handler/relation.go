package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/tolo-relation-stub/internal/store"
)

type RelationHandler struct {
	store *store.Store
}

func NewRelationHandler(s *store.Store) *RelationHandler {
	return &RelationHandler{store: s}
}

type eventRole struct {
	EventID string `json:"event_id"`
	Role    string `json:"role"`
}

type membership struct {
	UserID     string      `json:"user_id"`
	TenantID   string      `json:"tenant_id"`
	TenantRole string      `json:"tenant_role"`
	EventRoles []eventRole `json:"event_roles"`
}

type membershipsResponse struct {
	Memberships []membership `json:"memberships"`
}

func (h *RelationHandler) ListMemberships(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.Param("userId")

	if _, ok := h.store.Tenant(tenantID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	resp := membershipsResponse{Memberships: []membership{}}

	role, ok := h.store.TenantRole(tenantID, userID)
	if !ok {
		c.JSON(http.StatusOK, resp)
		return
	}

	m := membership{
		UserID:     userID,
		TenantID:   tenantID,
		TenantRole: string(role),
		EventRoles: []eventRole{},
	}
	for _, e := range h.store.EventsByTenant(tenantID) {
		role, ok := h.store.EventRole(e.ID, userID)
		if !ok {
			continue
		}
		m.EventRoles = append(m.EventRoles, eventRole{EventID: e.ID, Role: string(role)})
	}
	resp.Memberships = append(resp.Memberships, m)

	c.JSON(http.StatusOK, resp)
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
