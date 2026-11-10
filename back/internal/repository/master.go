package postgres

import (
	"context"
	"fmt"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type MasterRepo struct {
	db DB
}

func NewMasterRepo(db DB) (*MasterRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("master repository: db is nil")
	}
	return &MasterRepo{db: db}, nil
}

const masterColumns = `
	id,
	user_id,
	slug,
	title,
	category,
	COALESCE(description, ''),
	timezone,
	subscription_until,
	created_at,
	updated_at`

func (r *MasterRepo) CreateMaster(ctx context.Context, master domain.Master) (domain.Master, error) {
	query := `
		INSERT INTO masters (
			user_id, slug, title, category, description, timezone,
			subscription_until, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + masterColumns

	created, err := scanMaster(r.db.QueryRow(ctx, query,
		master.UserID,
		master.Slug,
		master.Title,
		master.Category,
		master.Description,
		master.Timezone,
		master.SubscriptionUntil,
		master.CreatedAt,
		master.UpdatedAt,
	))
	if err != nil {
		return domain.Master{}, fmt.Errorf("create master: %w", err)
	}
	return created, nil
}

func (r *MasterRepo) GetMasterByUserID(ctx context.Context, userID int64) (domain.Master, error) {
	query := `SELECT ` + masterColumns + ` FROM masters WHERE user_id = $1`
	master, err := scanMaster(r.db.QueryRow(ctx, query, userID))
	if err != nil {
		return domain.Master{}, fmt.Errorf("get master by user id: %w", err)
	}
	return master, nil
}

func (r *MasterRepo) GetPublicMasterBySlug(ctx context.Context, slug string) (domain.Master, error) {
	query := `SELECT ` + masterColumns + ` FROM masters WHERE slug = $1`
	master, err := scanMaster(r.db.QueryRow(ctx, query, slug))
	if err != nil {
		return domain.Master{}, fmt.Errorf("get public master by slug: %w", err)
	}
	return master, nil
}

func (r *MasterRepo) UpdateMaster(ctx context.Context, master domain.Master) (domain.Master, error) {
	query := `
		UPDATE masters
		SET title = $3,
		    category = $4,
		    description = $5,
		    timezone = $6,
		    updated_at = $7
		WHERE id = $1 AND user_id = $2
		RETURNING ` + masterColumns

	updated, err := scanMaster(r.db.QueryRow(ctx, query,
		master.ID,
		master.UserID,
		master.Title,
		master.Category,
		master.Description,
		master.Timezone,
		master.UpdatedAt,
	))
	if err != nil {
		return domain.Master{}, fmt.Errorf("update master: %w", err)
	}
	return updated, nil
}

const serviceColumns = `
	id,
	master_id,
	name,
	COALESCE(description, ''),
	duration_minutes,
	price,
	is_active,
	created_at,
	updated_at`

func (r *MasterRepo) ListServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error) {
	return r.listServices(ctx, masterID, false)
}

func (r *MasterRepo) ListActiveServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error) {
	return r.listServices(ctx, masterID, true)
}

func (r *MasterRepo) listServices(ctx context.Context, masterID int64, activeOnly bool) ([]domain.Service, error) {
	query := `SELECT ` + serviceColumns + ` FROM services WHERE master_id = $1`
	if activeOnly {
		query += ` AND is_active = TRUE`
	}
	query += ` ORDER BY created_at, id`

	rows, err := r.db.Query(ctx, query, masterID)
	if err != nil {
		return nil, fmt.Errorf("list services by master id: %w", err)
	}
	defer rows.Close()

	services := make([]domain.Service, 0)
	for rows.Next() {
		service, err := scanService(rows)
		if err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}
	return services, nil
}

func (r *MasterRepo) CreateService(ctx context.Context, service domain.Service) (domain.Service, error) {
	query := `
		INSERT INTO services (
			master_id, name, description, duration_minutes, price,
			is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + serviceColumns

	created, err := scanService(r.db.QueryRow(ctx, query,
		service.MasterID,
		service.Name,
		service.Description,
		service.DurationMinutes,
		service.Price,
		service.IsActive,
		service.CreatedAt,
		service.UpdatedAt,
	))
	if err != nil {
		return domain.Service{}, fmt.Errorf("create service: %w", err)
	}
	return created, nil
}

func (r *MasterRepo) GetServiceByIDAndMasterID(ctx context.Context, serviceID, masterID int64) (domain.Service, error) {
	query := `
		SELECT ` + serviceColumns + `
		FROM services
		WHERE id = $1 AND master_id = $2`

	service, err := scanService(r.db.QueryRow(ctx, query, serviceID, masterID))
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service by id and master id: %w", err)
	}
	return service, nil
}

func (r *MasterRepo) UpdateService(ctx context.Context, service domain.Service) (domain.Service, error) {
	query := `
		UPDATE services
		SET name = $3,
		    description = $4,
		    duration_minutes = $5,
		    price = $6,
		    is_active = $7,
		    updated_at = $8
		WHERE id = $1 AND master_id = $2
		RETURNING ` + serviceColumns

	updated, err := scanService(r.db.QueryRow(ctx, query,
		service.ID,
		service.MasterID,
		service.Name,
		service.Description,
		service.DurationMinutes,
		service.Price,
		service.IsActive,
		service.UpdatedAt,
	))
	if err != nil {
		return domain.Service{}, fmt.Errorf("update service: %w", err)
	}
	return updated, nil
}

const scheduleColumns = `
	id,
	master_id,
	day_of_week,
	start_time,
	end_time,
	created_at,
	updated_at`

func (r *MasterRepo) ReplaceSchedulePeriods(ctx context.Context, masterID int64, periods []domain.SchedulePeriod) ([]domain.SchedulePeriod, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin replace schedule transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err := tx.Exec(ctx, `DELETE FROM schedule_periods WHERE master_id = $1`, masterID); err != nil {
		return nil, fmt.Errorf("delete existing schedule periods: %w", err)
	}

	saved := make([]domain.SchedulePeriod, 0, len(periods))
	for _, period := range periods {
		query := `
			INSERT INTO schedule_periods (
				master_id, day_of_week, start_time, end_time, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING ` + scheduleColumns

		created, err := scanSchedulePeriod(tx.QueryRow(ctx, query,
			masterID,
			period.DayOfWeek,
			period.StartTime,
			period.EndTime,
			period.CreatedAt,
			period.UpdatedAt,
		))
		if err != nil {
			return nil, fmt.Errorf("insert schedule period: %w", err)
		}
		saved = append(saved, created)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit replace schedule transaction: %w", err)
	}
	committed = true
	return saved, nil
}

func (r *MasterRepo) ListSchedulePeriodsByMasterID(ctx context.Context, masterID int64) ([]domain.SchedulePeriod, error) {
	query := `
		SELECT ` + scheduleColumns + `
		FROM schedule_periods
		WHERE master_id = $1
		ORDER BY day_of_week, start_time, id`

	rows, err := r.db.Query(ctx, query, masterID)
	if err != nil {
		return nil, fmt.Errorf("list schedule periods by master id: %w", err)
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

func scanMaster(row pgx.Row) (domain.Master, error) {
	var master domain.Master
	err := row.Scan(
		&master.ID,
		&master.UserID,
		&master.Slug,
		&master.Title,
		&master.Category,
		&master.Description,
		&master.Timezone,
		&master.SubscriptionUntil,
		&master.CreatedAt,
		&master.UpdatedAt,
	)
	return master, err
}

func scanService(row pgx.Row) (domain.Service, error) {
	var service domain.Service
	err := row.Scan(
		&service.ID,
		&service.MasterID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.IsActive,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	return service, err
}

func scanSchedulePeriod(row pgx.Row) (domain.SchedulePeriod, error) {
	var period domain.SchedulePeriod
	err := row.Scan(
		&period.ID,
		&period.MasterID,
		&period.DayOfWeek,
		&period.StartTime,
		&period.EndTime,
		&period.CreatedAt,
		&period.UpdatedAt,
	)
	return period, err
}
