package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestUserRegistrationIPMigrationDefinesStorageAndIndexes(t *testing.T) {
	content, err := migrations.FS.ReadFile("194_user_registration_ip_controls.sql")
	require.NoError(t, err)
	sqlText := strings.ToLower(string(content))

	require.Contains(t, sqlText, "add column if not exists registration_ip inet")
	require.Contains(t, sqlText, "create table if not exists blocked_registration_ips")
	require.Contains(t, sqlText, "ip_address inet not null unique")
	require.Contains(t, sqlText, "idx_users_registration_ip")
}
