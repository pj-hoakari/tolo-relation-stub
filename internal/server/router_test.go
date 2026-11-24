package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
		nil,
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

func TestVerifyRegistrationGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := store.New(
		[]domain.Tenant{{ID: "tenant-a"}, {ID: "tenant-b"}},
		nil, nil, nil,
		[]domain.RegistrationGrant{
			{TenantID: "tenant-a", Type: domain.InvitationGrant, Token: "invite-a"},
			{TenantID: "tenant-a", Type: domain.OwnershipClaimGrant, Token: "claim-a"},
		},
	)
	router := NewRouter(s)

	const (
		malformed = `{"error":"request body is malformed"}`
		required  = `{"error":"exactly one of invitation_token and ownership_claim_token is required"}`
		valid     = `{"valid":true}`
		invalid   = `{"valid":false}`
	)

	tests := []struct {
		name       string
		tenant     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "invitation token", tenant: "tenant-a", body: `{"invitation_token":"invite-a"}`, wantStatus: http.StatusOK, wantBody: valid},
		{name: "ownership claim token", tenant: "tenant-a", body: `{"ownership_claim_token":"claim-a"}`, wantStatus: http.StatusOK, wantBody: valid},
		{name: "token under the other type", tenant: "tenant-a", body: `{"ownership_claim_token":"invite-a"}`, wantStatus: http.StatusOK, wantBody: invalid},
		{name: "token of another tenant", tenant: "tenant-b", body: `{"invitation_token":"invite-a"}`, wantStatus: http.StatusOK, wantBody: invalid},
		{name: "unknown tenant", tenant: "tenant-x", body: `{"invitation_token":"invite-a"}`, wantStatus: http.StatusOK, wantBody: invalid},
		{name: "unknown token", tenant: "tenant-a", body: `{"invitation_token":"nope"}`, wantStatus: http.StatusOK, wantBody: invalid},
		{name: "token kept as sent", tenant: "tenant-a", body: `{"invitation_token":" invite-a "}`, wantStatus: http.StatusOK, wantBody: invalid},
		{name: "empty body", tenant: "tenant-a", body: "", wantStatus: http.StatusBadRequest, wantBody: malformed},
		{name: "not JSON", tenant: "tenant-a", body: "invitation_token=invite-a", wantStatus: http.StatusBadRequest, wantBody: malformed},
		{name: "unknown field", tenant: "tenant-a", body: `{"invitation_token":"invite-a","tenant_id":"other"}`, wantStatus: http.StatusBadRequest, wantBody: malformed},
		{name: "trailing data", tenant: "tenant-a", body: `{"invitation_token":"invite-a"}{}`, wantStatus: http.StatusBadRequest, wantBody: malformed},
		{name: "more than 4 KiB", tenant: "tenant-a", body: `{"invitation_token":"` + strings.Repeat("a", 4<<10) + `"}`, wantStatus: http.StatusBadRequest, wantBody: malformed},
		{name: "both tokens", tenant: "tenant-a", body: `{"invitation_token":"invite-a","ownership_claim_token":"claim-a"}`, wantStatus: http.StatusBadRequest, wantBody: required},
		{name: "empty token beside a token", tenant: "tenant-a", body: `{"invitation_token":"","ownership_claim_token":"claim-a"}`, wantStatus: http.StatusBadRequest, wantBody: required},
		{name: "neither token", tenant: "tenant-a", body: `{}`, wantStatus: http.StatusBadRequest, wantBody: required},
		{name: "empty token", tenant: "tenant-a", body: `{"invitation_token":""}`, wantStatus: http.StatusBadRequest, wantBody: required},
		{name: "blank tenant", tenant: "%20%20%20", body: `{"invitation_token":"invite-a"}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"tenant_id is required"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			path := "/tenants/" + tt.tenant + "/registration-grants/verify"
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %s, want %s", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Fatalf("Cache-Control = %q, want no-store", got)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q, want JSON", got)
			}
		})
	}
}
