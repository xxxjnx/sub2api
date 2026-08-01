//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserRegistrationIPRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(integrationEntClient, integrationDB)
	ipRepo, ok := repo.(service.UserRegistrationIPRepository)
	require.True(t, ok)

	stamp := time.Now().UnixNano()
	ipAddress := fmt.Sprintf("2001:db8:%x:%x::1", uint64(stamp)>>48&0xffff, uint64(stamp)>>16&0xffff)
	newUser := func(suffix string) *service.User {
		user := &service.User{
			Email:          fmt.Sprintf("registration-ip-%d-%s@example.com", stamp, suffix),
			Role:           service.RoleUser,
			Status:         service.StatusActive,
			Concurrency:    1,
			RegistrationIP: ipAddress,
		}
		require.NoError(t, user.SetPassword("test-password"))
		return user
	}

	first := newUser("first")
	second := newUser("second")
	require.NoError(t, repo.Create(ctx, first))
	defer func() { _ = repo.Delete(ctx, first.ID) }()
	require.NoError(t, repo.Create(ctx, second))
	defer func() { _ = repo.Delete(ctx, second.ID) }()

	info, err := ipRepo.GetRegistrationIPInfoByUserIDs(ctx, []int64{first.ID, second.ID})
	require.NoError(t, err)
	require.Equal(t, ipAddress, info[first.ID].IPAddress)
	require.False(t, info[first.ID].Blocked)

	risks, total, err := ipRepo.ListRegistrationIPRisks(ctx, pagination.PaginationParams{Page: 1, PageSize: 100})
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(1))
	var found *service.RegistrationIPRisk
	for i := range risks {
		if risks[i].IPAddress == ipAddress {
			found = &risks[i]
			break
		}
	}
	require.NotNil(t, found)
	require.Equal(t, int64(2), found.UserCount)
	require.Len(t, found.Users, 2)

	block, err := ipRepo.BlockRegistrationIP(ctx, ipAddress, "duplicate registration test", first.ID)
	require.NoError(t, err)
	require.Equal(t, ipAddress, block.IPAddress)
	defer func() { _ = ipRepo.UnblockRegistrationIP(ctx, ipAddress) }()

	blocked, err := ipRepo.IsRegistrationIPBlocked(ctx, ipAddress)
	require.NoError(t, err)
	require.True(t, blocked)

	third := newUser("blocked")
	err = repo.Create(ctx, third)
	require.True(t, errors.Is(err, service.ErrRegistrationIPBlocked))

	require.NoError(t, ipRepo.UnblockRegistrationIP(ctx, ipAddress))
	third = newUser("unblocked")
	require.NoError(t, repo.Create(ctx, third))
	defer func() { _ = repo.Delete(ctx, third.ID) }()
}
