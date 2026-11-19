package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
)

type BookingRepo struct {
	db DB
}

func NewBookingRepo(db DB) (*BookingRepo, error) {
	if db == nil {
		return nil, errors.New("booking repository: db is nil")
	}
	return &BookingRepo{db: db}, nil
}

func (r *BookingRepo) GetMasterBySlug(ctx context.Context, slug string) (domain.Master, error) {
	query := `SELECT ` + masterColumns + ` FROM masters WHERE slug = $1`
	master, err := scanMaster(r.db.QueryRow(ctx, query, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Master{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Master{}, fmt.Errorf("get master by slug: %w", err)
	}
	return master, nil
}

func (r *BookingRepo) GetActiveServiceByID(
	ctx context.Context,
	masterID int64,
	serviceID int64,
) (domain.Service, error) {
	query := `
		SELECT ` + serviceColumns + `
		FROM services
		WHERE master_id = $1 AND id = $2 AND is_active = TRUE`
	service, err := scanService(r.db.QueryRow(ctx, query, masterID, serviceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Service{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Service{}, fmt.Errorf("get active service by id: %w", err)
	}
	return service, nil
}

func (r *BookingRepo) ListSchedulePeriodsByDay(
	ctx context.Context,
	masterID int64,
	dayOfWeek int8,
) ([]domain.SchedulePeriod, error) {
	query := `
		SELECT ` + scheduleColumns + `
		FROM schedule_periods
		WHERE master_id = $1 AND day_of_week = $2
		ORDER BY start_time, id`
	rows, err := r.db.Query(ctx, query, masterID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("list schedule periods by day: %w", err)
	}
	defer rows.Close()

	periods := make([]domain.SchedulePeriod, 0)
	for rows.Next() {
		period, err := scanSchedulePeriod(rows)
		if err != nil {
			return nil, fmt.Errorf("scan schedule period: %w", err)
		}
		periods = append(periods, period)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedule periods: %w", err)
	}
	return periods, nil
}

func (r *BookingRepo) ListActiveBookingsOverlapping(
	ctx context.Context,
	masterID int64,
	startAt time.Time,
	endAt time.Time,
) ([]domain.Booking, error) {
	query := `
		SELECT id, start_at, end_at
		FROM bookings
		WHERE master_id = $1
		  AND status IN ('booked', 'confirmed_by_client')
		  AND start_at < $3
		  AND end_at > $2
		ORDER BY start_at, id`
	rows, err := r.db.Query(ctx, query, masterID, startAt, endAt)
	if err != nil {
		return nil, fmt.Errorf("list active overlapping bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(&booking.ID, &booking.StartAt, &booking.EndAt); err != nil {
			return nil, fmt.Errorf("scan overlapping booking: %w", err)
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overlapping bookings: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepo) CreateBooking(ctx context.Context, booking domain.Booking) (domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Booking{}, fmt.Errorf("begin create booking transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedMasterID int64
	err = tx.QueryRow(ctx, `SELECT id FROM masters WHERE id = $1 FOR UPDATE`, booking.MasterID).Scan(&lockedMasterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Booking{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("lock master for booking: %w", err)
	}

	var activeService bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM services
			WHERE id = $1 AND master_id = $2 AND is_active = TRUE
		)
	`, booking.ServiceID, booking.MasterID).Scan(&activeService)
	if err != nil {
		return domain.Booking{}, fmt.Errorf("recheck active service: %w", err)
	}
	if !activeService {
		return domain.Booking{}, domain.ErrNotFound
	}

	var overlaps bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE master_id = $1
			  AND status IN ('booked', 'confirmed_by_client')
			  AND start_at < $3
			  AND end_at > $2
		)
	`, booking.MasterID, booking.StartAt, booking.EndAt).Scan(&overlaps)
	if err != nil {
		return domain.Booking{}, fmt.Errorf("recheck booking overlap: %w", err)
	}
	if overlaps {
		return domain.Booking{}, domain.ErrBookingConflict
	}

	query := `
		INSERT INTO bookings (
			master_id, client_id, service_id, start_at, end_at, status,
			client_comment, service_name_snapshot, duration_minutes_snapshot,
			price_snapshot, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'booked', NULLIF($6, ''), $7, $8, $9, $10, $11)
		RETURNING
			id, master_id, client_id, service_id, start_at, end_at, status,
			COALESCE(client_comment, ''), service_name_snapshot,
			duration_minutes_snapshot, price_snapshot, created_at, updated_at`
	created, err := scanBooking(tx.QueryRow(ctx, query,
		booking.MasterID,
		booking.ClientID,
		booking.ServiceID,
		booking.StartAt,
		booking.EndAt,
		booking.ClientComment,
		booking.ServiceNameSnapshot,
		booking.DurationMinutesSnapshot,
		booking.PriceSnapshot,
		booking.CreatedAt,
		booking.UpdatedAt,
	))
	if err != nil {
		return domain.Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Booking{}, fmt.Errorf("commit create booking transaction: %w", err)
	}
	committed = true
	return created, nil
}

func scanBooking(row pgx.Row) (domain.Booking, error) {
	var booking domain.Booking
	err := row.Scan(
		&booking.ID,
		&booking.MasterID,
		&booking.ClientID,
		&booking.ServiceID,
		&booking.StartAt,
		&booking.EndAt,
		&booking.Status,
		&booking.ClientComment,
		&booking.ServiceNameSnapshot,
		&booking.DurationMinutesSnapshot,
		&booking.PriceSnapshot,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	return booking, err
}
