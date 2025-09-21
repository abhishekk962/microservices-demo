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

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/microservices-demo/src/frontend/activitylog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (fe *frontendServer) listActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	
	// Parse query parameters
	limit := 100 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get activities
	activities, err := activitylog.GetRecentActivities(limit)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get activities"), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func (fe *frontendServer) sessionActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	sessionID := sessionID(r)

	// Parse query parameters
	limit := 50 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get session activities
	activities, err := activitylog.GetActivitiesBySession(sessionID, limit)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get session activities"), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func (fe *frontendServer) activityStatsHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)

	// Parse time range parameters
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // default to last 24 hours

	if start := r.URL.Query().Get("start"); start != "" {
		if st, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = st
		}
	}
	if end := r.URL.Query().Get("end"); end != "" {
		if et, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = et
		}
	}

	// Get activity statistics
	stats, err := activitylog.GetActivityStats(startTime, endTime)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get activity stats"), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}