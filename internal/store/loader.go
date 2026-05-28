package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KasumiMercury/tolo-relation-stub/internal/domain"
)

type seedFile struct {
	Tenants       []seedTenant           `json:"tenants"`
	Events        []seedEvent            `json:"events"`
	TenantMembers []seedTenantMembership `json:"tenantMembers"`
	EventMembers  []seedEventMembership  `json:"eventMembers"`
}

type seedTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type seedEvent struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
}

type seedTenantMembership struct {
	TenantID string `json:"tenantId"`
	UserID   string `json:"userId"`
	Role     string `json:"role"`
}

type seedEventMembership struct {
	EventID string `json:"eventId"`
	UserID  string `json:"userId"`
	Role    string `json:"role"`
}

func LoadFromFile(path string) (*Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open seed file: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	var raw seedFile
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode seed file: %w", err)
	}

	tenants := make([]domain.Tenant, 0, len(raw.Tenants))
	tenantIDs := make(map[string]struct{}, len(raw.Tenants))
	for i, t := range raw.Tenants {
		if t.ID == "" {
			return nil, fmt.Errorf("tenants[%d]: id is required", i)
		}
		if _, dup := tenantIDs[t.ID]; dup {
			return nil, fmt.Errorf("tenants[%d]: duplicate id %q", i, t.ID)
		}
		tenantIDs[t.ID] = struct{}{}
		tenants = append(tenants, domain.Tenant{ID: t.ID, Name: t.Name})
	}

	events := make([]domain.Event, 0, len(raw.Events))
	eventIDs := make(map[string]struct{}, len(raw.Events))
	for i, e := range raw.Events {
		if e.ID == "" {
			return nil, fmt.Errorf("events[%d]: id is required", i)
		}
		if _, dup := eventIDs[e.ID]; dup {
			return nil, fmt.Errorf("events[%d]: duplicate id %q", i, e.ID)
		}
		if _, ok := tenantIDs[e.TenantID]; !ok {
			return nil, fmt.Errorf("events[%d]: tenantId %q does not exist", i, e.TenantID)
		}
		eventIDs[e.ID] = struct{}{}
		events = append(events, domain.Event{ID: e.ID, TenantID: e.TenantID, Name: e.Name})
	}

	eventTenant := make(map[string]string, len(events))
	for _, e := range events {
		eventTenant[e.ID] = e.TenantID
	}

	tenantMembers := make([]domain.TenantMembership, 0, len(raw.TenantMembers))
	userTenant := make(map[string]string, len(raw.TenantMembers))
	for i, m := range raw.TenantMembers {
		if m.UserID == "" {
			return nil, fmt.Errorf("tenantMembers[%d]: userId is required", i)
		}
		if _, ok := tenantIDs[m.TenantID]; !ok {
			return nil, fmt.Errorf("tenantMembers[%d]: tenantId %q does not exist", i, m.TenantID)
		}
		if existing, dup := userTenant[m.UserID]; dup {
			return nil, fmt.Errorf("tenantMembers[%d]: user %q already belongs to tenant %q (a user must belong to exactly one tenant)", i, m.UserID, existing)
		}
		role := domain.Role(m.Role)
		if !role.Valid() {
			return nil, fmt.Errorf("tenantMembers[%d]: invalid role %q", i, m.Role)
		}
		userTenant[m.UserID] = m.TenantID
		tenantMembers = append(tenantMembers, domain.TenantMembership{
			TenantID: m.TenantID, UserID: m.UserID, Role: role,
		})
	}

	eventMembers := make([]domain.EventMembership, 0, len(raw.EventMembers))
	for i, m := range raw.EventMembers {
		if m.UserID == "" {
			return nil, fmt.Errorf("eventMembers[%d]: userId is required", i)
		}
		tenantID, ok := eventTenant[m.EventID]
		if !ok {
			return nil, fmt.Errorf("eventMembers[%d]: eventId %q does not exist", i, m.EventID)
		}
		userTenantID, ok := userTenant[m.UserID]
		if !ok {
			return nil, fmt.Errorf("eventMembers[%d]: user %q does not belong to any tenant (a user must have a tenant role to hold an event role)", i, m.UserID)
		}
		if userTenantID != tenantID {
			return nil, fmt.Errorf("eventMembers[%d]: user %q belongs to tenant %q but event %q belongs to tenant %q", i, m.UserID, userTenantID, m.EventID, tenantID)
		}
		role := domain.Role(m.Role)
		if !role.Valid() {
			return nil, fmt.Errorf("eventMembers[%d]: invalid role %q", i, m.Role)
		}
		eventMembers = append(eventMembers, domain.EventMembership{
			EventID: m.EventID, UserID: m.UserID, Role: role,
		})
	}

	return New(tenants, events, tenantMembers, eventMembers), nil
}
