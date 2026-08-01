package service

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrRegistrationIPBlocked = infraerrors.Forbidden(
		"REGISTRATION_IP_BLOCKED",
		"registration is not allowed from this IP address",
	)
	ErrRegistrationIPInvalid = infraerrors.BadRequest(
		"REGISTRATION_IP_INVALID",
		"invalid IP address",
	)
)

// RegistrationIPInfo is the administrator-only registration network data for
// a user. IPAddress is empty for accounts created before this data was tracked
// or accounts created by an administrator.
type RegistrationIPInfo struct {
	IPAddress string
	Blocked   bool
}

type RegistrationIPRiskUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type RegistrationIPRisk struct {
	IPAddress         string                   `json:"ip_address"`
	UserCount         int64                    `json:"user_count"`
	Users             []RegistrationIPRiskUser `json:"users"`
	FirstRegisteredAt *time.Time               `json:"first_registered_at,omitempty"`
	LastRegisteredAt  *time.Time               `json:"last_registered_at,omitempty"`
	Blocked           bool                     `json:"blocked"`
	BlockReason       string                   `json:"block_reason"`
	BlockedAt         *time.Time               `json:"blocked_at,omitempty"`
	BlockedBy         *int64                   `json:"blocked_by,omitempty"`
}

type RegistrationIPBlock struct {
	ID        int64     `json:"id"`
	IPAddress string    `json:"ip_address"`
	Reason    string    `json:"reason"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRegistrationIPRepository owns registration-IP persistence, duplicate
// registration reporting, and the block list. It is separate from
// UserRepository so existing narrow test doubles and integrations do not gain
// unrelated methods.
type UserRegistrationIPRepository interface {
	GetRegistrationIPInfoByUserIDs(ctx context.Context, userIDs []int64) (map[int64]RegistrationIPInfo, error)
	ListRegistrationIPRisks(ctx context.Context, params pagination.PaginationParams) ([]RegistrationIPRisk, int64, error)
	IsRegistrationIPBlocked(ctx context.Context, ipAddress string) (bool, error)
	BlockRegistrationIP(ctx context.Context, ipAddress, reason string, actorAdminID int64) (*RegistrationIPBlock, error)
	UnblockRegistrationIP(ctx context.Context, ipAddress string) error
}

// NormalizeRegistrationIP converts IPv4-mapped IPv6 values to IPv4 and rejects
// zone-scoped addresses, which PostgreSQL INET cannot store consistently.
func NormalizeRegistrationIP(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	addr, err := netip.ParseAddr(value)
	if err != nil || addr.Zone() != "" {
		return "", ErrRegistrationIPInvalid
	}
	return addr.Unmap().String(), nil
}

func registrationIPFromContext(ctx context.Context) (string, error) {
	binding := SessionBindingFromContext(ctx)
	if binding == nil {
		return "", nil
	}
	return NormalizeRegistrationIP(binding.IP)
}

// prepareRegistrationUser is called by every public account-creation path. The
// repository repeats the block check while holding the per-IP transaction lock
// so an administrator blocking an IP cannot race a concurrent registration.
func (s *AuthService) prepareRegistrationUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("prepare registration user: user is nil")
	}
	ipAddress, err := registrationIPFromContext(ctx)
	if err != nil {
		return ErrServiceUnavailable
	}
	if ipAddress == "" {
		return nil
	}

	repo, ok := s.userRepo.(UserRegistrationIPRepository)
	if !ok {
		return ErrServiceUnavailable
	}
	blocked, err := repo.IsRegistrationIPBlocked(ctx, ipAddress)
	if err != nil {
		return ErrServiceUnavailable
	}
	if blocked {
		return ErrRegistrationIPBlocked
	}

	user.RegistrationIP = ipAddress
	return nil
}
