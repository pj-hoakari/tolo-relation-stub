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

type tenantBlock struct {
	ID   string  `json:"id"`
	Role *string `json:"role"`
}

type userBlock struct {
	ID string `json:"id"`
}

type eventBlock struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type relationResponse struct {
	Tenant tenantBlock  `json:"tenant"`
	User   userBlock    `json:"user"`
	Events []eventBlock `json:"events"`
}

func (h *RelationHandler) GetTenantUserRelation(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.Param("userId")

	if _, ok := h.store.Tenant(tenantID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	resp := relationResponse{
		Tenant: tenantBlock{ID: tenantID},
		User:   userBlock{ID: userID},
		Events: []eventBlock{},
	}

	if role, ok := h.store.TenantRole(tenantID, userID); ok {
		r := string(role)
		resp.Tenant.Role = &r
	}

	for _, e := range h.store.EventsByTenant(tenantID) {
		role, ok := h.store.EventRole(e.ID, userID)
		if !ok {
			continue
		}
		resp.Events = append(resp.Events, eventBlock{ID: e.ID, Role: string(role)})
	}

	c.JSON(http.StatusOK, resp)
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
