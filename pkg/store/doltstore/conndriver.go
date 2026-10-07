package doltstore

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

// connDriver is an Ent dialect.Driver over a single, caller-owned *sql.Conn.
//
// Ent's stock entsql.Driver can only begin transactions on a *sql.DB (it
// type-asserts its ExecQuerier), so a store built over a *sql.Conn panicked on
// the first operation that opens a transaction — which includes every Ent
// update. connDriver begins transactions with (*sql.Conn).BeginTx instead, so
// they run on that same physical connection and see any session state the
// caller set on it beforehand (for example a tenant GUC that row-level
// security policies read).
//
// Close is a no-op: the caller owns the connection's lifecycle.
type connDriver struct {
	*entsql.Driver
	conn *sql.Conn
}

var _ dialect.Driver = (*connDriver)(nil)

func newConnDriver(dialectName string, conn *sql.Conn) *connDriver {
	return &connDriver{
		Driver: entsql.NewDriver(dialectName, entsql.Conn{ExecQuerier: conn}),
		conn:   conn,
	}
}

// Tx starts a transaction on the underlying connection.
func (d *connDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	return d.BeginTx(ctx, nil)
}

// BeginTx starts a transaction with options on the underlying connection.
func (d *connDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	tx, err := d.conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("begin transaction on connection: %w", err)
	}
	return &entsql.Tx{Conn: entsql.Conn{ExecQuerier: tx}, Tx: tx}, nil
}

// Close does not close the connection; the caller owns it.
func (d *connDriver) Close() error { return nil }
