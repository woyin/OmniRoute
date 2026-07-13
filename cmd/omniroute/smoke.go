package main

import (
	"fmt"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func sqliteSmoke(cfg *config.Config) error {
	dbConn, err := db.OpenDB(cfg)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("begin smoke transaction: %w", err)
	}
	defer tx.Rollback()

	const value = "ok"
	if _, err := tx.Exec(`
		INSERT INTO key_value (namespace, key, value) VALUES (?, ?, ?)
		ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value`,
		"smoke", "release", value); err != nil {
		return fmt.Errorf("write smoke value: %w", err)
	}

	var got string
	if err := tx.QueryRow(
		"SELECT value FROM key_value WHERE namespace = ? AND key = ?",
		"smoke", "release",
	).Scan(&got); err != nil {
		return fmt.Errorf("read smoke value: %w", err)
	}
	if got != value {
		return fmt.Errorf("verify smoke value: got %q, want %q", got, value)
	}
	return nil
}
