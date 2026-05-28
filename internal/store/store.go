package store

import (
	"github.com/KasumiMercury/tolo-relation-stub/internal/domain"
)

type Store struct {
	tenants        map[string]domain.Tenant
	eventsByTenant map[string][]domain.Event
	tenantRoles    map[string]domain.Role
	eventRoles     map[string]domain.Role
}

func New(
	tenants []domain.Tenant,
	events []domain.Event,
	tenantMembers []domain.TenantMembership,
	eventMembers []domain.EventMembership,
) *Store {
	s := &Store{
		tenants:        make(map[string]domain.Tenant, len(tenants)),
		eventsByTenant: make(map[string][]domain.Event),
		tenantRoles:    make(map[string]domain.Role, len(tenantMembers)),
		eventRoles:     make(map[string]domain.Role, len(eventMembers)),
	}
	for _, t := range tenants {
		s.tenants[t.ID] = t
	}
	for _, e := range events {
		s.eventsByTenant[e.TenantID] = append(s.eventsByTenant[e.TenantID], e)
	}
	for _, m := range tenantMembers {
		s.tenantRoles[membershipKey(m.TenantID, m.UserID)] = m.Role
	}
	for _, m := range eventMembers {
		s.eventRoles[membershipKey(m.EventID, m.UserID)] = m.Role
	}
	return s
}

func (s *Store) Tenant(id string) (domain.Tenant, bool) {
	t, ok := s.tenants[id]
	return t, ok
}

func (s *Store) EventsByTenant(tenantID string) []domain.Event {
	return s.eventsByTenant[tenantID]
}

func (s *Store) TenantRole(tenantID, userID string) (domain.Role, bool) {
	r, ok := s.tenantRoles[membershipKey(tenantID, userID)]
	return r, ok
}

func (s *Store) EventRole(eventID, userID string) (domain.Role, bool) {
	r, ok := s.eventRoles[membershipKey(eventID, userID)]
	return r, ok
}

func membershipKey(resourceID, userID string) string {
	return resourceID + "\x00" + userID
}
