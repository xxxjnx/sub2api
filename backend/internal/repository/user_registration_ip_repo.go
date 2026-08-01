package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

var _ service.UserRegistrationIPRepository = (*userRepository)(nil)

const registrationIPRiskCTE = `
WITH candidate_ips AS (
    SELECT registration_ip AS ip_address
    FROM users
    WHERE registration_ip IS NOT NULL AND deleted_at IS NULL
    UNION
    SELECT ip_address
    FROM blocked_registration_ips
), risk_rows AS (
    SELECT
        candidates.ip_address,
        COUNT(users.id)::BIGINT AS user_count,
        MIN(users.created_at) AS first_registered_at,
        MAX(users.created_at) AS last_registered_at,
        (blocks.id IS NOT NULL) AS blocked,
        COALESCE(blocks.reason, '') AS block_reason,
        blocks.created_at AS blocked_at,
        blocks.created_by AS blocked_by
    FROM candidate_ips AS candidates
    LEFT JOIN users
        ON users.registration_ip = candidates.ip_address
       AND users.deleted_at IS NULL
    LEFT JOIN blocked_registration_ips AS blocks
        ON blocks.ip_address = candidates.ip_address
    GROUP BY
        candidates.ip_address,
        blocks.id,
        blocks.reason,
        blocks.created_at,
        blocks.created_by
)
`

func registrationIPLockKey(ipAddress string) string {
	return "registration-ip:" + ipAddress
}

func registrationIPBlockedWithExecutor(ctx context.Context, exec sqlQueryExecutor, ipAddress string) (bool, error) {
	if exec == nil {
		return false, fmt.Errorf("registration IP repository: sql executor is not configured")
	}
	rows, err := exec.QueryContext(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM blocked_registration_ips
            WHERE ip_address = $1
        )
    `, ipAddress)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if rowsErr := rows.Err(); rowsErr != nil {
			return false, rowsErr
		}
		return false, sql.ErrNoRows
	}
	var blocked bool
	if err := rows.Scan(&blocked); err != nil {
		return false, err
	}
	return blocked, rows.Err()
}

func (r *userRepository) IsRegistrationIPBlocked(ctx context.Context, ipAddress string) (bool, error) {
	normalized, err := service.NormalizeRegistrationIP(ipAddress)
	if err != nil {
		return false, err
	}
	if normalized == "" {
		return false, nil
	}
	return registrationIPBlockedWithExecutor(ctx, txAwareSQLExecutor(ctx, r.sql, r.client), normalized)
}

func (r *userRepository) GetRegistrationIPInfoByUserIDs(ctx context.Context, userIDs []int64) (map[int64]service.RegistrationIPInfo, error) {
	result := make(map[int64]service.RegistrationIPInfo, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	if r.sql == nil {
		return nil, fmt.Errorf("registration IP repository: sql executor is not configured")
	}

	rows, err := r.sql.QueryContext(ctx, `
        SELECT
            users.id,
            COALESCE(host(users.registration_ip), ''),
            (blocks.id IS NOT NULL)
        FROM users
        LEFT JOIN blocked_registration_ips AS blocks
            ON blocks.ip_address = users.registration_ip
        WHERE users.id = ANY($1)
    `, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			userID int64
			info   service.RegistrationIPInfo
		)
		if err := rows.Scan(&userID, &info.IPAddress, &info.Blocked); err != nil {
			return nil, err
		}
		result[userID] = info
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepository) ListRegistrationIPRisks(ctx context.Context, params pagination.PaginationParams) ([]service.RegistrationIPRisk, int64, error) {
	if r.sql == nil {
		return nil, 0, fmt.Errorf("registration IP repository: sql executor is not configured")
	}

	countRows, err := r.sql.QueryContext(ctx, registrationIPRiskCTE+`
        SELECT COUNT(*)
        FROM risk_rows
        WHERE user_count > 1 OR blocked
    `)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if countRows.Next() {
		err = countRows.Scan(&total)
	} else if rowsErr := countRows.Err(); rowsErr != nil {
		err = rowsErr
	} else {
		err = sql.ErrNoRows
	}
	_ = countRows.Close()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.sql.QueryContext(ctx, registrationIPRiskCTE+`
        SELECT
            host(ip_address),
            user_count,
            first_registered_at,
            last_registered_at,
            blocked,
            block_reason,
            blocked_at,
            blocked_by
        FROM risk_rows
        WHERE user_count > 1 OR blocked
        ORDER BY blocked DESC, user_count DESC, last_registered_at DESC NULLS LAST, ip_address
        OFFSET $1
        LIMIT $2
    `, params.Offset(), params.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	risks := make([]service.RegistrationIPRisk, 0, params.Limit())
	for rows.Next() {
		var (
			risk            service.RegistrationIPRisk
			firstRegistered sql.NullTime
			lastRegistered  sql.NullTime
			blockedAt       sql.NullTime
			blockedBy       sql.NullInt64
		)
		if err := rows.Scan(
			&risk.IPAddress,
			&risk.UserCount,
			&firstRegistered,
			&lastRegistered,
			&risk.Blocked,
			&risk.BlockReason,
			&blockedAt,
			&blockedBy,
		); err != nil {
			return nil, 0, err
		}
		if firstRegistered.Valid {
			risk.FirstRegisteredAt = &firstRegistered.Time
		}
		if lastRegistered.Valid {
			risk.LastRegisteredAt = &lastRegistered.Time
		}
		if blockedAt.Valid {
			risk.BlockedAt = &blockedAt.Time
		}
		if blockedBy.Valid {
			risk.BlockedBy = &blockedBy.Int64
		}
		risk.Users = make([]service.RegistrationIPRiskUser, 0, risk.UserCount)
		risks = append(risks, risk)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := r.loadRegistrationIPRiskUsers(ctx, risks); err != nil {
		return nil, 0, err
	}
	return risks, total, nil
}

func (r *userRepository) loadRegistrationIPRiskUsers(ctx context.Context, risks []service.RegistrationIPRisk) error {
	if len(risks) == 0 {
		return nil
	}
	ipAddresses := make([]string, 0, len(risks))
	byIP := make(map[string]*service.RegistrationIPRisk, len(risks))
	for i := range risks {
		ipAddresses = append(ipAddresses, risks[i].IPAddress)
		byIP[risks[i].IPAddress] = &risks[i]
	}

	rows, err := r.sql.QueryContext(ctx, `
        SELECT
            host(registration_ip),
            id,
            email,
            username,
            status,
            created_at
        FROM users
        WHERE deleted_at IS NULL
          AND registration_ip = ANY($1::inet[])
        ORDER BY created_at ASC, id ASC
    `, pq.Array(ipAddresses))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			ipAddress string
			user      service.RegistrationIPRiskUser
		)
		if err := rows.Scan(
			&ipAddress,
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Status,
			&user.CreatedAt,
		); err != nil {
			return err
		}
		if risk := byIP[ipAddress]; risk != nil {
			risk.Users = append(risk.Users, user)
		}
	}
	return rows.Err()
}

func (r *userRepository) BlockRegistrationIP(ctx context.Context, ipAddress, reason string, actorAdminID int64) (*service.RegistrationIPBlock, error) {
	normalized, err := service.NormalizeRegistrationIP(ipAddress)
	if err != nil {
		return nil, err
	}
	if normalized == "" {
		return nil, service.ErrRegistrationIPInvalid
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	exec := txAwareSQLExecutor(txCtx, r.sql, r.client)
	release, err := lockRepositoryScopedKeys(txCtx, tx.Client(), exec, registrationIPLockKey(normalized))
	if err != nil {
		return nil, err
	}
	defer release()

	createdBy := sql.NullInt64{Int64: actorAdminID, Valid: actorAdminID > 0}
	rows, err := exec.QueryContext(txCtx, `
        INSERT INTO blocked_registration_ips (ip_address, reason, created_by)
        VALUES ($1, $2, $3)
        ON CONFLICT (ip_address) DO UPDATE SET
            reason = EXCLUDED.reason,
            created_by = EXCLUDED.created_by,
            updated_at = NOW()
        RETURNING id, host(ip_address), reason, created_by, created_at, updated_at
    `, normalized, strings.TrimSpace(reason), createdBy)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if rowsErr := rows.Err(); rowsErr != nil {
			return nil, rowsErr
		}
		return nil, sql.ErrNoRows
	}

	var (
		block        service.RegistrationIPBlock
		blockCreator sql.NullInt64
	)
	if err := rows.Scan(
		&block.ID,
		&block.IPAddress,
		&block.Reason,
		&blockCreator,
		&block.CreatedAt,
		&block.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if blockCreator.Valid {
		block.CreatedBy = &blockCreator.Int64
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *userRepository) UnblockRegistrationIP(ctx context.Context, ipAddress string) error {
	normalized, err := service.NormalizeRegistrationIP(ipAddress)
	if err != nil {
		return err
	}
	if normalized == "" {
		return service.ErrRegistrationIPInvalid
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	exec := txAwareSQLExecutor(txCtx, r.sql, r.client)
	release, err := lockRepositoryScopedKeys(txCtx, tx.Client(), exec, registrationIPLockKey(normalized))
	if err != nil {
		return err
	}
	defer release()

	if _, err := exec.ExecContext(txCtx, "DELETE FROM blocked_registration_ips WHERE ip_address = $1", normalized); err != nil {
		return err
	}
	return tx.Commit()
}
