package doltstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// postgresTestDSNEnv names the environment variable holding an admin DSN for
// a Postgres server the connection-scoped store tests may create scratch
// databases on. The tests skip when it is unset or the server is unreachable.
const postgresTestDSNEnv = "VISIONSTUDIO_TEST_POSTGRES_DSN"

// newScratchPostgres creates a throwaway database on the server named by
// VISIONSTUDIO_TEST_POSTGRES_DSN, migrates the schema into it, and returns a
// pooled handle to it. The database is dropped on cleanup.
func newScratchPostgres(t *testing.T) *sql.DB {
	t.Helper()
	adminDSN := os.Getenv(postgresTestDSNEnv)
	if adminDSN == "" {
		t.Skipf("%s not set", postgresTestDSNEnv)
	}
	ctx := context.Background()
	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("open admin: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Close(); err != nil {
			t.Errorf("close admin: %v", err)
		}
	})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := admin.PingContext(pingCtx); err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	name := "vs_conntest_" + hex.EncodeToString(buf)
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create scratch database: %v", err)
	}

	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	u.Path = "/" + name
	pooled, err := NewPostgres(u.String())
	if err != nil {
		t.Fatalf("open scratch: %v", err)
	}
	t.Cleanup(func() {
		if err := pooled.Close(); err != nil {
			t.Errorf("close scratch: %v", err)
		}
		if _, err := admin.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop scratch database: %v", err)
		}
	})
	if err := pooled.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pooled.DB()
}

// tenantConn acquires a dedicated connection and sets a session-level GUC on
// it, the way the cloud scopes a request to a tenant.
func tenantConn(t *testing.T, db *sql.DB, tenant string) *sql.Conn {
	t.Helper()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Errorf("release conn: %v", err)
		}
	})
	if _, err := conn.ExecContext(ctx, "SELECT set_config('app.tenant', $1, false)", tenant); err != nil {
		t.Fatalf("set tenant: %v", err)
	}
	return conn
}

func TestNewPostgresFromConnSupportsUpdatesAndTransactions(t *testing.T) {
	db := newScratchPostgres(t)
	ctx := context.Background()
	conn := tenantConn(t, db, "acme")
	ds := NewPostgresFromConn(conn)
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)

	in := &store.Initiative{ID: "INIT-APP-001", Organization: "default", Title: "One", Status: "draft", CreatedAt: now, UpdatedAt: now}
	if err := ds.CreateInitiative(ctx, in); err != nil {
		t.Fatalf("create initiative: %v", err)
	}

	// Ent opens a transaction for every update; over a *sql.Conn this used
	// to panic with an interface conversion to *sql.DB.
	in.Title = "One, revised"
	in.Status = "executing"
	in.UpdatedAt = now.Add(time.Hour)
	if err := ds.UpdateInitiative(ctx, in); err != nil {
		t.Fatalf("update initiative: %v", err)
	}
	got, err := ds.GetInitiative(ctx, in.ID)
	if err != nil {
		t.Fatalf("get initiative: %v", err)
	}
	if got.Title != "One, revised" || got.Status != "executing" {
		t.Fatalf("update not applied: %+v", got)
	}

	repo := &store.Repository{ID: "github.com/testorg/app", Organization: "testorg", RepositoryName: "app", DefaultBranch: "main", Status: "active"}
	if err := ds.CreateRepository(ctx, repo); err != nil {
		t.Fatalf("create repository: %v", err)
	}
	rmi := &store.RoadmapItem{ID: "RMI-APP-001", RepositoryID: repo.ID, InitiativeID: in.ID, Title: "x", ItemType: "capability", Status: "planned", CreatedAt: now, UpdatedAt: now}
	if err := ds.CreateRMI(ctx, rmi); err != nil {
		t.Fatalf("create rmi: %v", err)
	}
	rmi.Status = "in_progress"
	if err := ds.UpdateRMI(ctx, rmi); err != nil {
		t.Fatalf("update rmi: %v", err)
	}

	t.Run("explicit ent transaction commits and rolls back", func(t *testing.T) {
		tx, err := ds.Client().Tx(ctx)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := tx.Initiative.UpdateOneID(in.ID).SetTitle("committed").Save(ctx); err != nil {
			t.Fatalf("update in tx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		tx, err = ds.Client().Tx(ctx)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := tx.Initiative.UpdateOneID(in.ID).SetTitle("rolled back").Save(ctx); err != nil {
			t.Fatalf("update in tx: %v", err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatalf("rollback: %v", err)
		}
		got, err := ds.GetInitiative(ctx, in.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Title != "committed" {
			t.Fatalf("title = %q, want committed", got.Title)
		}
	})

	t.Run("unit of work runs without a Dolt commit", func(t *testing.T) {
		uow := NewUnitOfWork(ds, "test")
		err := uow.Execute(ctx, func(ctx context.Context, s store.Store) error {
			return s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-APP-002", Organization: "default", Title: "Two", Status: "draft", CreatedAt: now, UpdatedAt: now})
		})
		if err != nil {
			t.Fatalf("unit of work: %v", err)
		}
		if _, err := ds.GetInitiative(ctx, "INIT-APP-002"); err != nil {
			t.Fatalf("get after unit of work: %v", err)
		}
	})

	t.Run("close leaves the caller's connection open", func(t *testing.T) {
		if err := ds.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		if err := conn.PingContext(ctx); err != nil {
			t.Fatalf("conn closed by store: %v", err)
		}
	})
}

func TestConnDriverTransactionSeesSessionGUC(t *testing.T) {
	db := newScratchPostgres(t)
	ctx := context.Background()
	conn := tenantConn(t, db, "tenant-42")
	drv := newConnDriver(dialect.Postgres, conn)

	tx, err := drv.Tx(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	var rows entsql.Rows
	if err := tx.Query(ctx, "SELECT current_setting('app.tenant'), pg_backend_pid()", []any{}, &rows); err != nil {
		t.Fatalf("query in tx: %v", err)
	}
	var tenant string
	var txPID int
	if !rows.Next() {
		t.Fatal("no row")
	}
	if err := rows.Scan(&tenant, &txPID); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if tenant != "tenant-42" {
		t.Fatalf("tenant inside tx = %q, want tenant-42", tenant)
	}

	var connPID int
	if err := conn.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&connPID); err != nil {
		t.Fatal(err)
	}
	if connPID != txPID {
		t.Fatalf("transaction ran on backend %d, connection is %d", txPID, connPID)
	}
	if !strings.HasPrefix(drv.Dialect(), dialect.Postgres) {
		t.Fatalf("dialect = %q", drv.Dialect())
	}
}
