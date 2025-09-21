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
	"encoding/json"
	"time"
)

// Common activity types
const (
	ActivityTypePageView      = "page_view"
	ActivityTypeAddToCart     = "add_to_cart"
	ActivityTypeEmptyCart     = "empty_cart"
	ActivityTypeCheckout      = "checkout"
	ActivityTypeCurrencyChange = "currency_change"
	ActivityTypeProductView   = "product_view"
)

// LogActivity records a new activity in the database
func LogActivity(activity *ActivityLog) error {
	query := `
		INSERT INTO activities (
			session_id, request_id, activity_type, path, method, 
			status_code, user_currency, details, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := GetDB().Exec(
		query,
		activity.SessionID,
		activity.RequestID,
		activity.ActivityType,
		activity.Path,
		activity.Method,
		activity.StatusCode,
		activity.UserCurrency,
		activity.Details,
		time.Now(),
	)
	return err
}

// GetActivitiesBySession retrieves all activities for a given session
func GetActivitiesBySession(sessionID string, limit int) ([]ActivityLog, error) {
	query := `
		SELECT id, session_id, request_id, activity_type, path, method,
			   status_code, user_currency, details, created_at
		FROM activities 
		WHERE session_id = ?
		ORDER BY created_at DESC
		LIMIT ?`

	return queryActivities(query, sessionID, limit)
}

// GetRecentActivities retrieves recent activities across all sessions
func GetRecentActivities(limit int) ([]ActivityLog, error) {
	query := `
		SELECT id, session_id, request_id, activity_type, path, method,
			   status_code, user_currency, details, created_at
		FROM activities
		ORDER BY created_at DESC
		LIMIT ?`

	return queryActivities(query, limit)
}

// GetActivityStats returns activity statistics for a given time period
func GetActivityStats(startTime, endTime time.Time) (map[string]int, error) {
	query := `
		SELECT activity_type, COUNT(*) as count
		FROM activities
		WHERE created_at BETWEEN ? AND ?
		GROUP BY activity_type`

	rows, err := GetDB().Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var activityType string
		var count int
		if err := rows.Scan(&activityType, &count); err != nil {
			return nil, err
		}
		stats[activityType] = count
	}
	return stats, rows.Err()
}

// Helper function to query and scan activities
func queryActivities(query string, args ...interface{}) ([]ActivityLog, error) {
	rows, err := GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []ActivityLog
	for rows.Next() {
		var activity ActivityLog
		err := rows.Scan(
			&activity.ID,
			&activity.SessionID,
			&activity.RequestID,
			&activity.ActivityType,
			&activity.Path,
			&activity.Method,
			&activity.StatusCode,
			&activity.UserCurrency,
			&activity.Details,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}