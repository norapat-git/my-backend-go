package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/sijms/go-ora/v2"
)

var (
	db   *sql.DB
	once sync.Once
)

// GetDB returns a singleton Oracle DB connection pool.
func GetDB() *sql.DB {
	once.Do(func() {
		user := os.Getenv("NODE_ORACLEDB_USER")
		password := os.Getenv("NODE_ORACLEDB_PASSWORD")
		connStr := os.Getenv("NODE_ORACLEDB_CONNECTIONSTRING")

		// go-ora connection string format: oracle://user:password@host:port/service
		dsn := fmt.Sprintf("oracle://%s:%s@%s", user, password, connStr)

		var err error
		db, err = sql.Open("oracle", dsn)
		if err != nil {
			log.Fatalf("Failed to open Oracle DB: %v", err)
		}

		// Pool settings (equivalent to poolMax: 200, queueMax: 500)
		db.SetMaxOpenConns(200)
		db.SetMaxIdleConns(10)

		if err = db.Ping(); err != nil {
			log.Printf("[WARNING] Oracle DB ping failed (will retry on first request): %v", err)
		} else {
			log.Println("[DB] Oracle connection pool initialized successfully")
		}
	})
	return db
}

// CloseDB closes the database connection pool.
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
