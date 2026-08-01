//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type authRegistrationIPRepoStub struct {
	UserRepository
	blocked    bool
	blockedErr error
	checkedIP  string
}

func (s *authRegistrationIPRepoStub) GetRegistrationIPInfoByUserIDs(context.Context, []int64) (map[int64]RegistrationIPInfo, error) {
	panic("unexpected GetRegistrationIPInfoByUserIDs call")
}

func (s *authRegistrationIPRepoStub) ListRegistrationIPRisks(context.Context, pagination.PaginationParams) ([]RegistrationIPRisk, int64, error) {
	panic("unexpected ListRegistrationIPRisks call")
}

func (s *authRegistrationIPRepoStub) IsRegistrationIPBlocked(_ context.Context, ipAddress string) (bool, error) {
	s.checkedIP = ipAddress
	return s.blocked, s.blockedErr
}

func (s *authRegistrationIPRepoStub) BlockRegistrationIP(context.Context, string, string, int64) (*RegistrationIPBlock, error) {
	panic("unexpected BlockRegistrationIP call")
}

func (s *authRegistrationIPRepoStub) UnblockRegistrationIP(context.Context, string) error {
	panic("unexpected UnblockRegistrationIP call")
}

func TestNormalizeRegistrationIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   bool
	}{
		{name: "ipv4", input: " 203.0.113.7 ", want: "203.0.113.7"},
		{name: "ipv4 mapped ipv6", input: "::ffff:203.0.113.7", want: "203.0.113.7"},
		{name: "ipv6", input: "2001:0db8::1", want: "2001:db8::1"},
		{name: "empty", input: "", want: ""},
		{name: "invalid", input: "not-an-ip", err: true},
		{name: "zone scoped", input: "fe80::1%eth0", err: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeRegistrationIP(test.input)
			if test.err {
				require.ErrorIs(t, err, ErrRegistrationIPInvalid)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestPrepareRegistrationUserCapturesCanonicalIP(t *testing.T) {
	repo := &authRegistrationIPRepoStub{}
	svc := &AuthService{userRepo: repo}
	ctx := WithSessionBinding(context.Background(), &SessionBinding{IP: "::ffff:203.0.113.7"})
	user := &User{}

	err := svc.prepareRegistrationUser(ctx, user)

	require.NoError(t, err)
	require.Equal(t, "203.0.113.7", repo.checkedIP)
	require.Equal(t, "203.0.113.7", user.RegistrationIP)
}

func TestPrepareRegistrationUserRejectsBlockedIP(t *testing.T) {
	repo := &authRegistrationIPRepoStub{blocked: true}
	svc := &AuthService{userRepo: repo}
	ctx := WithSessionBinding(context.Background(), &SessionBinding{IP: "2001:db8::9"})

	err := svc.prepareRegistrationUser(ctx, &User{})

	require.ErrorIs(t, err, ErrRegistrationIPBlocked)
}

func TestPrepareRegistrationUserFailsClosedWhenBlockLookupFails(t *testing.T) {
	repo := &authRegistrationIPRepoStub{blockedErr: errors.New("database unavailable")}
	svc := &AuthService{userRepo: repo}
	ctx := WithSessionBinding(context.Background(), &SessionBinding{IP: "203.0.113.8"})

	err := svc.prepareRegistrationUser(ctx, &User{})

	require.ErrorIs(t, err, ErrServiceUnavailable)
}
