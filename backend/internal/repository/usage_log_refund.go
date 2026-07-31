package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// RefundUsage atomically marks a usage charge as refunded and reverses the
// corresponding wallet or subscription charge. The row lock is the
// idempotency boundary for concurrent refund attempts.
func (r *usageLogRepository) RefundUsage(ctx context.Context, input service.RefundUsageInput) (_ int64, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		userID         int64
		billingType    int16
		subscriptionID sql.NullInt64
		actualCost     float64
		refundAmount   float64
		refundedAt     sql.NullTime
		createdAt      time.Time
	)
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, billing_type, subscription_id, actual_cost,
		       refund_amount, refunded_at, created_at
		FROM usage_logs
		WHERE id = $1
		FOR UPDATE
	`, input.UsageID).Scan(
		&userID, &billingType, &subscriptionID, &actualCost,
		&refundAmount, &refundedAt, &createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUsageLogNotFound
	}
	if err != nil {
		return 0, err
	}
	if refundedAt.Valid || refundAmount > 0 {
		return 0, service.ErrUsageAlreadyRefunded
	}
	if actualCost <= 0 {
		return 0, service.ErrUsageNotRefundable
	}

	switch int8(billingType) {
	case service.BillingTypeBalance:
		result, execErr := tx.ExecContext(ctx, `
			UPDATE users
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
		`, actualCost, userID)
		if execErr != nil {
			return 0, execErr
		}
		if err := requireOneRefundRow(result); err != nil {
			return 0, err
		}
	case service.BillingTypeSubscription:
		if !subscriptionID.Valid {
			return 0, service.ErrUsageRefundIncomplete
		}
		result, execErr := tx.ExecContext(ctx, `
			UPDATE user_subscriptions
			SET daily_usage_usd = CASE
					WHEN daily_window_start IS NULL OR daily_window_start <= $1
					THEN GREATEST(0, daily_usage_usd - $2)
					ELSE daily_usage_usd
				END,
				weekly_usage_usd = CASE
					WHEN weekly_window_start IS NULL OR weekly_window_start <= $1
					THEN GREATEST(0, weekly_usage_usd - $2)
					ELSE weekly_usage_usd
				END,
				monthly_usage_usd = CASE
					WHEN monthly_window_start IS NULL OR monthly_window_start <= $1
					THEN GREATEST(0, monthly_usage_usd - $2)
					ELSE monthly_usage_usd
				END,
				updated_at = NOW()
			WHERE id = $3 AND user_id = $4
		`, createdAt, actualCost, subscriptionID.Int64, userID)
		if execErr != nil {
			return 0, execErr
		}
		if err := requireOneRefundRow(result); err != nil {
			return 0, err
		}
	default:
		return 0, service.ErrUsageRefundIncomplete
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE usage_logs
		SET actual_cost = 0,
		    refund_amount = $1,
		    refund_reason = $2,
		    refunded_at = NOW(),
		    refunded_by = $3
		WHERE id = $4
	`, actualCost, input.Reason, input.AdminID, input.UsageID)
	if err != nil {
		return 0, err
	}
	if err := requireOneRefundRow(result); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	tx = nil
	return userID, nil
}

func requireOneRefundRow(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("%w: expected one billing row, got %d", service.ErrUsageRefundIncomplete, rows)
	}
	return nil
}
