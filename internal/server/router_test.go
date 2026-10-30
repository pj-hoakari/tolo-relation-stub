package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/tolo-relation-stub/internal/domain"
	"github.com/KasumiMercury/tolo-relation-stub/internal/store"
)

func TestListMemberships(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := store.New(
		[]domain.Tenant{{ID: "tenant-a"}, {ID: "tenant-b"}},
		[]domain.Event{{ID: "event-1", TenantID: "tenant-a"}, {ID: "event-2", TenantID: "tenant-a"}, {ID: "event-b1", TenantID: "tenant-b"}},
		[]domain.TenantMembership{{TenantID: "tenant-a", UserID: "user-1", Role: domain.RoleStaff}},
		[]domain.EventMembership{{EventID: "event-2", UserID: "user-1", Role: domain.RoleOwner}},
	)
	router := NewRouter(s)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "member",
			path:       "/tenants/tenant-a/users/user-1/memberships",
			wantStatus: http.StatusOK,
			wantBody:   `{"memberships":[{"user_id":"user-1","tenant_id":"tenant-a","tenant_role":"staff","event_roles":[{"event_id":"event-2","role":"owner"}]}]}`,
		},
		{
			name:       "not a member of an existing tenant",
			path:       "/tenants/tenant-b/users/user-1/memberships",
			wantStatus: http.StatusOK,
			wantBody:   `{"memberships":[]}`,
		},
		{
			name:       "unknown tenant",
			path:       "/tenants/tenant-x/users/user-1/memberships",
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"tenant not found"}`,
		},
		{
			name:       "legacy route is gone",
			path:       "/tenants/tenant-a/users/user-1",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %s, want %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
