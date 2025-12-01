package repository

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	"github.com/jackc/pgx/v5"
)

type txKeyType string

const txKey txKeyType = "tx"

type DeliveryPostgresRepository struct {
	DB *db.Database
}

func NewDeliveryPostgresRepository(db *db.Database) *DeliveryPostgresRepository {
	return &DeliveryPostgresRepository{DB: db}
}

func getTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

func (r *DeliveryPostgresRepository) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *DeliveryPostgresRepository) Create(ctx context.Context, d *model.Delivery) error {
	const query = `
        INSERT INTO delivery (courier_id, order_id, assigned_at, deadline)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `

	tx, ok := getTx(ctx)
	if ok {
		return tx.QueryRow(ctx, query,
			d.CourierID, d.OrderID, d.AssignedAt, d.Deadline,
		).Scan(&d.ID)
	}

	return r.DB.Pool.QueryRow(ctx, query,
		d.CourierID, d.OrderID, d.AssignedAt, d.Deadline,
	).Scan(&d.ID)
}

func (r *DeliveryPostgresRepository) DeleteByOrderID(ctx context.Context, orderID string) (*model.Delivery, error) {
	const query = `
        DELETE FROM delivery 
        WHERE order_id=$1 
        RETURNING id, courier_id, order_id, assigned_at, deadline;
    `

	d := &model.Delivery{}
	tx, ok := getTx(ctx)

	var err error
	if ok {
		err = tx.QueryRow(ctx, query, orderID).Scan(
			&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline,
		)
	} else {
		err = r.DB.Pool.QueryRow(ctx, query, orderID).Scan(
			&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline,
		)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return d, nil
}

func (r *DeliveryPostgresRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error) {
	const query = `
        SELECT id, courier_id, order_id, assigned_at, deadline
        FROM delivery
        WHERE order_id=$1;
    `

	d := &model.Delivery{}
	tx, ok := getTx(ctx)

	var err error
	if ok {
		err = tx.QueryRow(ctx, query, orderID).Scan(
			&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline,
		)
	} else {
		err = r.DB.Pool.QueryRow(ctx, query, orderID).Scan(
			&d.ID, &d.CourierID, &d.OrderID, &d.AssignedAt, &d.Deadline,
		)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return d, nil
}
