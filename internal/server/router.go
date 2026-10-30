package server

import (
	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/tolo-relation-stub/internal/handler"
	"github.com/KasumiMercury/tolo-relation-stub/internal/store"
)

func NewRouter(s *store.Store) *gin.Engine {
	r := gin.Default()

	rel := handler.NewRelationHandler(s)

	r.GET("/healthz", handler.Healthz)
	r.GET("/tenants/:tenantId/users/:userId/memberships", rel.ListMemberships)

	return r
}
