package repository

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type txKeyType string

const txKey txKeyType = "tx"

type PostgresCourierRepository struct {
	DB *db.Database
}

func NewPostgresCourierRepository(db *db.Database) *PostgresCourierRepository {
	return &PostgresCourierRepository{DB: db}
}

func getTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

func (r *PostgresCourierRepository) Create(ctx context.Context, c *model.Courier) error {
	query := `
        INSERT INTO couriers (name, phone, status, transport_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `

	err := r.DB.Pool.QueryRow(ctx, query,
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

func (r *PostgresCourierRepository) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	c := &model.Courier{}

	query := `
        SELECT id, name, phone, status, transport_type, created_at, updated_at
        FROM couriers WHERE id=$1;
    `

	err := r.DB.Pool.QueryRow(ctx, query, id).Scan(
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

func (r *PostgresCourierRepository) GetAll(ctx context.Context) ([]*model.Courier, error) {
	query := `
        SELECT id, name, phone, status, transport_type
        FROM couriers;
    `

	rows, err := r.DB.Pool.Query(ctx, query)
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

func (r *PostgresCourierRepository) Update(ctx context.Context, c *model.Courier) error {
	query := `
        UPDATE couriers SET 
            name = COALESCE($2, name),
            phone = COALESCE($3, phone),
            status = COALESCE($4, status),
            transport_type = COALESCE($5, transport_type),
            updated_at = now()
        WHERE id = $1;
    `

	cmd, err := r.DB.Pool.Exec(ctx, query,
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

func (r *PostgresCourierRepository) FindAvailable(ctx context.Context) (*model.Courier, error) {
	query := `
        SELECT id, name, phone, status, transport_type
        FROM couriers
        WHERE status = 'available'
        ORDER BY id LIMIT 1;
    `

	c := &model.Courier{}

	err := r.DB.Pool.QueryRow(ctx, query).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

func (r *PostgresCourierRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE couriers SET status=$2 WHERE id=$1;`

	tx, ok := getTx(ctx)
	if ok {
		_, err := tx.Exec(ctx, query, id, status)
		return err
	}

	_, err := r.DB.Pool.Exec(ctx, query, id, status)
	return err
}
