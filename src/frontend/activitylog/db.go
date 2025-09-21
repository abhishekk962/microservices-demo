// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activitylog

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const (
	dbFileName = "activities.db"
	schema = `
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		request_id TEXT NOT NULL,
		activity_type TEXT NOT NULL,
		path TEXT NOT NULL,
		method TEXT NOT NULL,
		status_code INTEGER,
		user_currency TEXT,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_session ON activities(session_id);
	CREATE INDEX IF NOT EXISTS idx_created_at ON activities(created_at);
	CREATE INDEX IF NOT EXISTS idx_activity_type ON activities(activity_type);
	`
)

var (
	db   *sql.DB
	once sync.Once
)

// ActivityLog represents a single activity entry
type ActivityLog struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id"`
	RequestID    string    `json:"request_id"`
	ActivityType string    `json:"activity_type"`
	Path         string    `json:"path"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	UserCurrency string    `json:"user_currency"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `json:"created_at"`
}

// InitDB initializes the SQLite database connection and creates the schema
func InitDB(log logrus.FieldLogger) error {
	var err error
	once.Do(func() {
		// Create data directory if it doesn't exist
		dataDir := "data"
		if err = os.MkdirAll(dataDir, 0755); err != nil {
			return
		}

		dbPath := filepath.Join(dataDir, dbFileName)
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return
		}

		// Enable WAL mode for better concurrent performance
		if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return
		}

		// Create tables
		if _, err = db.Exec(schema); err != nil {
			return
		}

		log.Infof("Activity logging database initialized at: %s", dbPath)
	})
	return err
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}