package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KasumiMercury/tolo-relation-stub/internal/domain"
)

const maxRegistrationGrantBytes = 4 << 10

var (
	errGrantBodyMalformed = errors.New("request body is malformed")
	errGrantRequired      = errors.New("exactly one of invitation_token and ownership_claim_token is required")
)

type registrationGrantRequest struct {
	InvitationToken     *string `json:"invitation_token"`
	OwnershipClaimToken *string `json:"ownership_claim_token"`
}

type registrationGrantResponse struct {
	Valid bool `json:"valid"`
}

func (h *RelationHandler) VerifyRegistrationGrant(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	tenantID := strings.TrimSpace(c.Param("tenantId"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	grantType, token, err := parseRegistrationGrant(http.MaxBytesReader(c.Writer, c.Request.Body, maxRegistrationGrantBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid := h.store.HasRegistrationGrant(domain.RegistrationGrant{TenantID: tenantID, Type: grantType, Token: token})
	c.JSON(http.StatusOK, registrationGrantResponse{Valid: valid})
}

func parseRegistrationGrant(body io.Reader) (domain.RegistrationGrantType, string, error) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	var request registrationGrantRequest
	if err := decoder.Decode(&request); err != nil {
		return "", "", errGrantBodyMalformed
	}

	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return "", "", errGrantBodyMalformed
	}

	switch {
	case request.InvitationToken != nil && request.OwnershipClaimToken == nil && *request.InvitationToken != "":
		return domain.InvitationGrant, *request.InvitationToken, nil
	case request.OwnershipClaimToken != nil && request.InvitationToken == nil && *request.OwnershipClaimToken != "":
		return domain.OwnershipClaimGrant, *request.OwnershipClaimToken, nil
	default:
		return "", "", errGrantRequired
	}
}
