package domain

type Role string

const (
	RoleOwner Role = "owner"
	RoleStaff Role = "staff"
)

func (r Role) Valid() bool {
	return r == RoleOwner || r == RoleStaff
}

type Tenant struct {
	ID   string
	Name string
}

type Event struct {
	ID       string
	TenantID string
	Name     string
}

type TenantMembership struct {
	TenantID string
	UserID   string
	Role     Role
}

type EventMembership struct {
	EventID string
	UserID  string
	Role    Role
}

type RegistrationGrantType string

const (
	InvitationGrant     RegistrationGrantType = "invitation_token"
	OwnershipClaimGrant RegistrationGrantType = "ownership_claim_token"
)

func (t RegistrationGrantType) Valid() bool {
	return t == InvitationGrant || t == OwnershipClaimGrant
}

type RegistrationGrant struct {
	TenantID string
	Type     RegistrationGrantType
	Token    string
}
