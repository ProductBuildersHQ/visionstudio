package cloudsync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/grokify/godolt"
)

func TestPushRequiresTenantAndURL(t *testing.T) {
	if _, err := Push(context.Background(), nil, "", "", "file:///tmp/x"); err == nil {
		t.Fatal("expected error for missing tenant slug")
	}
	if _, err := Push(context.Background(), nil, "", "acme", ""); err == nil {
		t.Fatal("expected error for missing remote URL")
	}
}

// startServer mirrors the godolt/doltstore harness pattern: a throwaway
// dolt sql-server for fast integration testing. Skips when dolt is
// unavailable.
func startServer(t *testing.T) (*sql.DB, string) {
	t.Helper()
	if !godolt.Available() {
		t.Skip("dolt binary not available")
	}
	root := t.TempDir()
	dbDir := filepath.Join(root, "syncdb")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := godolt.InitDir(context.Background(), dbDir, "t", "t@local"); err != nil {
		t.Fatal(err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	// #nosec G204 -- test harness; port from net.Listen on loopback.
	srv := exec.Command("dolt", "sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port))
	srv.Dir = root
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = srv.Process.Kill()
		_, _ = srv.Process.Wait()
	})

	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/syncdb", port)
	var db *sql.DB
	deadline := time.Now().Add(15 * time.Second)
	for {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			pingErr := db.PingContext(ctx)
			cancel()
			if pingErr == nil {
				break
			}
			_ = db.Close()
			err = pingErr
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not ready: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, root
}

func TestPushIdempotentRemote(t *testing.T) {
	db, root := startServer(t)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CALL DOLT_COMMIT('--allow-empty','-m','seed')"); err != nil {
		t.Fatal(err)
	}

	remoteDir := filepath.Join(root, "remote")
	if err := os.MkdirAll(remoteDir, 0o755); err != nil {
		t.Fatal(err)
	}
	remoteURL := "file://" + remoteDir

	res1, err := Push(ctx, db, "", "acme", remoteURL)
	if err != nil {
		t.Fatalf("first push: %v", err)
	}
	if res1.RemoteName != "cloud-acme" || !res1.DogfoodOnly {
		t.Fatalf("result: %+v", res1)
	}

	// Second push must reuse the remote, not error on a duplicate add.
	res2, err := Push(ctx, db, "", "acme", remoteURL)
	if err != nil {
		t.Fatalf("second push (idempotency): %v", err)
	}
	if res2.Message == "" {
		t.Fatal("expected a status message on re-push")
	}

	c := godolt.New(db)
	remotes, err := c.Remotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(remotes) != 1 {
		t.Fatalf("remotes = %v, want exactly 1 (no duplicate add)", remotes)
	}
}

func TestPullRequiresTenantAndURL(t *testing.T) {
	if _, err := Pull(context.Background(), nil, "", "", "file:///tmp/x"); err == nil {
		t.Fatal("expected error for missing tenant slug")
	}
	if _, err := Pull(context.Background(), nil, "", "acme", ""); err == nil {
		t.Fatal("expected error for missing remote URL")
	}
}

// TestPushRejectsNonFastForwardThenPullResolves is the RMI-VISIONSTUDIO-535
// acceptance scenario end to end: two independent clones each commit
// before either syncs the other's push (true history divergence, not
// just an unpushed local commit), so the second push is genuinely
// non-fast-forward and Dolt's own remote logic rejects it. Push must
// translate that rejection into ErrNonFastForward; Pull must then merge
// cleanly and unblock a subsequent push.
func TestPushRejectsNonFastForwardThenPullResolves(t *testing.T) {
	dbA, root := startServer(t)
	ctx := context.Background()

	if _, err := dbA.ExecContext(ctx, "CALL DOLT_COMMIT('--allow-empty','-m','seed')"); err != nil {
		t.Fatal(err)
	}
	remoteDir := filepath.Join(root, "remote")
	if err := os.MkdirAll(remoteDir, 0o755); err != nil {
		t.Fatal(err)
	}
	remoteURL := "file://" + remoteDir

	if _, err := Push(ctx, dbA, "", "acme", remoteURL); err != nil {
		t.Fatalf("A initial push: %v", err)
	}

	// B clones at "seed" — same base as A, then the two diverge.
	cloneDir := filepath.Join(root, "dbB")
	if err := godolt.Clone(ctx, remoteURL, cloneDir); err != nil {
		t.Fatalf("clone: %v", err)
	}
	dbB, cleanup := startServerOn(t, cloneDir)
	defer cleanup()

	if _, err := dbA.ExecContext(ctx, "CALL DOLT_COMMIT('--allow-empty','-m','A-only')"); err != nil {
		t.Fatal(err)
	}
	if _, err := Push(ctx, dbA, "", "acme", remoteURL); err != nil {
		t.Fatalf("A second push: %v", err)
	}
	if _, err := dbB.ExecContext(ctx, "CALL DOLT_COMMIT('--allow-empty','-m','B-only')"); err != nil {
		t.Fatal(err)
	}

	// B's push must be rejected with our translated, actionable error —
	// not a raw Dolt error string.
	_, err := Push(ctx, dbB, "", "acme", remoteURL)
	if err == nil {
		t.Fatal("expected B's push to be rejected as non-fast-forward")
	}
	if !errors.Is(err, ErrNonFastForward) {
		t.Fatalf("push error = %v, want it to wrap ErrNonFastForward", err)
	}

	res, err := Pull(ctx, dbB, "", "acme", remoteURL)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if res.Conflicts != 0 {
		t.Fatalf("conflicts = %d, want 0 (both commits were --allow-empty)", res.Conflicts)
	}

	if _, err := Push(ctx, dbB, "", "acme", remoteURL); err != nil {
		t.Fatalf("push after pull should now succeed: %v", err)
	}
}

// startServerOn boots a throwaway dolt sql-server over an existing
// (already-initialized) Dolt data directory, e.g. a fresh Clone target.
func startServerOn(t *testing.T, dbDir string) (*sql.DB, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	// #nosec G204 -- test harness; port from net.Listen on loopback.
	srv := exec.Command("dolt", "sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port))
	srv.Dir = dbDir
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}

	dbName := filepath.Base(dbDir)
	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/%s", port, dbName)
	var db *sql.DB
	deadline := time.Now().Add(15 * time.Second)
	for {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			pingErr := db.PingContext(ctx)
			cancel()
			if pingErr == nil {
				break
			}
			_ = db.Close()
			err = pingErr
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not ready: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return db, func() {
		_ = db.Close()
		_ = srv.Process.Kill()
		_, _ = srv.Process.Wait()
	}
}
