package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KasumiMercury/tolo-relation-stub/internal/domain"
)

func loadSeed(t *testing.T, seed string) (*Store, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "seed.json")
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	return LoadFromFile(path)
}

func TestLoadRegistrationGrants(t *testing.T) {
	grant := domain.RegistrationGrant{TenantID: "tenant-a", Type: domain.InvitationGrant, Token: "invite-a"}

	s, err := loadSeed(t, `{"tenants":[{"id":"tenant-a"}]}`)
	if err != nil {
		t.Fatalf("seed without registrationGrants: %v", err)
	}
	if s.HasRegistrationGrant(grant) {
		t.Fatal("seed without registrationGrants holds a grant")
	}

	s, err = loadSeed(t, `{"tenants":[{"id":"tenant-a"}],"registrationGrants":[{"tenantId":"tenant-a","type":"invitation_token","token":"invite-a"}]}`)
	if err != nil {
		t.Fatalf("seed with registrationGrants: %v", err)
	}
	if !s.HasRegistrationGrant(grant) {
		t.Fatal("seeded grant is not held")
	}

	refused := map[string]string{
		"unknown tenant": `{"tenantId":"tenant-x","type":"invitation_token","token":"t"}`,
		"invalid type":   `{"tenantId":"tenant-a","type":"invite","token":"t"}`,
		"empty token":    `{"tenantId":"tenant-a","type":"ownership_claim_token","token":""}`,
	}
	for name, g := range refused {
		t.Run(name, func(t *testing.T) {
			_, err := loadSeed(t, `{"tenants":[{"id":"tenant-a"}],"registrationGrants":[`+g+`]}`)
			if err == nil || !strings.Contains(err.Error(), "registrationGrants[0]") {
				t.Fatalf("err = %v, want a registrationGrants[0] error", err)
			}
		})
	}
}
