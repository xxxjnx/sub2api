package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryRefundUsageBalance(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &usageLogRepository{db: db}
	createdAt := time.Now().Add(-time.Minute)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT user_id, billing_type, subscription_id, actual_cost,[\s\S]+FOR UPDATE`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "billing_type", "subscription_id", "actual_cost",
			"refund_amount", "refunded_at", "created_at",
		}).AddRow(int64(7), int16(service.BillingTypeBalance), nil, 1.25, 0, nil, createdAt))
	mock.ExpectExec(`UPDATE users[\s\S]+balance = balance \+ \$1`).
		WithArgs(1.25, int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE usage_logs[\s\S]+refund_amount = \$1`).
		WithArgs(1.25, "billing error", int64(9), int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	userID, err := repo.RefundUsage(context.Background(), service.RefundUsageInput{
		UsageID: 42,
		AdminID: 9,
		Reason:  "billing error",
	})
	require.NoError(t, err)
	require.Equal(t, int64(7), userID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryRefundUsageRejectsSecondRefund(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &usageLogRepository{db: db}
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "billing_type", "subscription_id", "actual_cost",
			"refund_amount", "refunded_at", "created_at",
		}).AddRow(int64(7), int16(service.BillingTypeBalance), nil, 0, 1.25, now, now.Add(-time.Minute)))
	mock.ExpectRollback()

	_, err = repo.RefundUsage(context.Background(), service.RefundUsageInput{UsageID: 42, AdminID: 9, Reason: "again"})
	require.ErrorIs(t, err, service.ErrUsageAlreadyRefunded)
	require.NoError(t, mock.ExpectationsWereMet())
}
