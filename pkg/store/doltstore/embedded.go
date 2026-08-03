//go:build dolt_embedded

package doltstore

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/dolthub/driver"

	"github.com/ProductBuildersHQ/visionstudio/ent"
)

const defaultDBName = "visionstudio"

// NewEmbedded creates a DoltStore using the embedded Dolt engine.
// dataDir is the parent directory for Dolt databases; the database
// subdirectory is created automatically if it does not exist.
// No external dolt binary or server process is required.
func NewEmbedded(dataDir string) (*DoltStore, error) {
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	params := url.Values{}
	params.Set("commitname", "visionstudio")
	params.Set("commitemail", "visionstudio@local")
	params.Set("database", defaultDBName)

	dsn := fmt.Sprintf("file://%s?%s", absDir, params.Encode())

	db, err := sql.Open("dolt", dsn)
	if err != nil {
		return nil, fmt.Errorf("open embedded dolt: %w", err)
	}

	if err := ensureDatabase(db); err != nil {
		if cerr := db.Close(); cerr != nil {
			return nil, fmt.Errorf("ensure database: %w (close also failed: %v)", err, cerr)
		}
		return nil, err
	}

	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	return &DoltStore{client: client, db: db}, nil
}

// ensureDatabase creates the database if it doesn't already exist.
func ensureDatabase(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(),
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", defaultDBName))
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	return nil
}
