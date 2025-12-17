package repository

import (
	"context"
	"time"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type txKeyType string

const txKey txKeyType = "tx"

type postgresCourierRepository struct {
	db *db.Database
}

var _ CourierRepository = (*postgresCourierRepository)(nil)

func NewCourierRepository(dbConn *db.Database) CourierRepository {
	return &postgresCourierRepository{db: dbConn}
}

func getTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

func (r *postgresCourierRepository) Create(ctx context.Context, c *model.Courier) error {
	query := `
		INSERT INTO couriers (name, phone, status, transport_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.db.Pool.QueryRow(ctx, query,
		c.Name, c.Phone, c.Status, c.TransportType,
	).Scan(&c.ID)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (r *postgresCourierRepository) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	c := &model.Courier{}

	query := `
		SELECT id, name, phone, status, transport_type, created_at, updated_at
		FROM couriers WHERE id=$1;
	`

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return c, nil
}

func (r *postgresCourierRepository) GetAll(ctx context.Context) ([]*model.Courier, error) {
	query := `
		SELECT id, name, phone, status, transport_type
		FROM couriers;
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.Courier, 0)

	for rows.Next() {
		c := &model.Courier{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

func (r *postgresCourierRepository) Update(ctx context.Context, c *model.Courier) error {
	query := `
		UPDATE couriers SET 
			name = COALESCE($2, name),
			phone = COALESCE($3, phone),
			status = COALESCE($4, status),
			transport_type = COALESCE($5, transport_type),
			updated_at = now()
		WHERE id = $1;
	`

	cmd, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.Name, c.Phone, c.Status, c.TransportType,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *postgresCourierRepository) FindAvailable(ctx context.Context) (*model.Courier, error) {
	query := `
		SELECT c.id, c.name, c.phone, c.status, c.transport_type
		FROM couriers c
		LEFT JOIN delivery d ON d.courier_id = c.id
		WHERE c.status = 'available'
		GROUP BY c.id, c.name, c.phone, c.status, c.transport_type
		ORDER BY COUNT(d.id) ASC, c.id ASC
		LIMIT 1;
	`

	c := &model.Courier{}

	err := r.db.Pool.QueryRow(ctx, query).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

func (r *postgresCourierRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE couriers SET status=$2 WHERE id=$1;`

	if tx, ok := getTx(ctx); ok {
		_, err := tx.Exec(ctx, query, id, status)
		return err
	}

	_, err := r.db.Pool.Exec(ctx, query, id, status)
	return err
}

func (r *postgresCourierRepository) ReleaseExpired(ctx context.Context, now time.Time) (int64, error) {
	const query = `
		UPDATE couriers c
		SET status = 'available'
		FROM delivery d
		WHERE c.id = d.courier_id
		  AND c.status = 'busy'
		  AND d.deadline < $1;
	`

	cmd, err := r.db.Pool.Exec(ctx, query, now)
	if err != nil {
		return 0, err
	}

	return cmd.RowsAffected(), nil
}
