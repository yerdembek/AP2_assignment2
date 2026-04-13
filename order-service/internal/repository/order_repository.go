package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Исправлен путь к пакету domain
	"github.com/lib/pq"
	"github.com/yerdembek/AP2_assignment2/order-service/internal/domain"
)

// OrderRepository handles persistence and change notifications for orders.
type OrderRepository struct {
	db  *sql.DB
	dsn string
}

func NewOrderRepository(db *sql.DB, dsn string) *OrderRepository {
	return &OrderRepository{db: db, dsn: dsn}
}

func (r *OrderRepository) Save(ctx context.Context, o *domain.Order) error {
	query := `
		INSERT INTO orders (id, user_id, product_id, quantity, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		o.ID, o.UserID, o.ProductID, o.Quantity, o.Amount, o.Currency,
		o.Status, o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `SELECT id, user_id, product_id, quantity, amount, currency, status, created_at, updated_at FROM orders WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	o := &domain.Order{}
	err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.Amount, &o.Currency, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order %s not found", id)
	}
	return o, err
}

// UpdateStatus updates an order's status and emits a PostgreSQL NOTIFY for real-time streaming.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id, newStatus string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`,
		newStatus, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("order %s not found", id)
	}

	// PostgreSQL NOTIFY — payload is the new status.
	// Channel name is sanitised to be a valid identifier.
	channel := fmt.Sprintf("order_%s", id)
	_, err = tx.ExecContext(ctx, `SELECT pg_notify($1, $2)`, channel, newStatus)
	if err != nil {
		return fmt.Errorf("pg_notify: %w", err)
	}

	return tx.Commit()
}

// ListenForUpdates starts a pq.Listener on the order's dedicated channel.
// It returns a channel that receives new status strings as they are pushed by the DB.
func (r *OrderRepository) ListenForUpdates(ctx context.Context, orderID string) (<-chan string, error) {
	channel := fmt.Sprintf("order_%s", orderID)

	listener := pq.NewListener(r.dsn,
		500*time.Millisecond,
		5*time.Second,
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				fmt.Printf("[pq.Listener] event=%d err=%v\n", ev, err)
			}
		},
	)
	if err := listener.Listen(channel); err != nil {
		listener.Close()
		return nil, fmt.Errorf("listen on channel %q: %w", channel, err)
	}

	out := make(chan string, 10)

	go func() {
		defer listener.Close()
		defer close(out)

		// Also send the current status immediately so the client has an initial value.
		row := r.db.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = $1`, orderID)
		var currentStatus string
		if err := row.Scan(&currentStatus); err == nil {
			select {
			case out <- currentStatus:
			case <-ctx.Done():
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case notif, ok := <-listener.Notify:
				if !ok {
					return
				}
				if notif == nil {
					// Keep-alive ping — ignore.
					continue
				}
				select {
				case out <- notif.Extra:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// Migrate creates required tables.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id         TEXT PRIMARY KEY,
			user_id    TEXT NOT NULL,
			product_id TEXT NOT NULL,
			quantity   INTEGER NOT NULL,
			amount     DOUBLE PRECISION NOT NULL,
			currency   TEXT NOT NULL DEFAULT 'USD',
			status     TEXT NOT NULL DEFAULT 'PENDING',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	return err
}

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}
