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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/GoogleCloudPlatform/microservices-demo/src/frontend"
)

// ActivityMiddleware wraps an http.Handler and logs activities
type ActivityMiddleware struct {
	log  logrus.FieldLogger
	next http.Handler
}

// NewActivityMiddleware creates a new activity logging middleware
func NewActivityMiddleware(log logrus.FieldLogger, next http.Handler) *ActivityMiddleware {
	return &ActivityMiddleware{
		log:  log,
		next: next,
	}
}

// getActivityType determines the type of activity based on the request
func getActivityType(r *http.Request) string {
	// Get the route pattern from mux router
	route := mux.CurrentRoute(r)
	if route == nil {
		return "unknown"
	}

	path, _ := route.GetPathTemplate()
	method := r.Method

	switch {
	case path == "/" && method == "GET":
		return ActivityTypePageView
	case path == "/cart" && method == "POST":
		return ActivityTypeAddToCart
	case path == "/cart/empty" && method == "POST":
		return ActivityTypeEmptyCart
	case path == "/cart/checkout" && method == "POST":
		return ActivityTypeCheckout
	case path == "/setCurrency" && method == "POST":
		return ActivityTypeCurrencyChange
	case strings.HasPrefix(path, "/product/") && method == "GET":
		return ActivityTypeProductView
	default:
		return "other"
	}
}

func (m *ActivityMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rr := &responseRecorder{w: w}

	// Extract common fields
	sessionID := r.Context().Value(ctxKeySessionID{}).(string)
	requestID := r.Context().Value(ctxKeyRequestID{}).(string)
	userCurrency := currentCurrency(r)

	// Create the activity log entry
	activity := &ActivityLog{
		SessionID:    sessionID,
		RequestID:    requestID,
		ActivityType: getActivityType(r),
		Path:        r.URL.Path,
		Method:      r.Method,
		UserCurrency: userCurrency,
	}

	// Call the next handler
	m.next.ServeHTTP(rr, r)

	// Record the response status
	activity.StatusCode = rr.status

	// Add any relevant details based on the activity type
	details := make(map[string]interface{})
	switch activity.ActivityType {
	case ActivityTypeAddToCart:
		details["product_id"] = r.FormValue("product_id")
		details["quantity"] = r.FormValue("quantity")
	case ActivityTypeProductView:
		if route := mux.CurrentRoute(r); route != nil {
			vars := mux.Vars(r)
			details["product_id"] = vars["id"]
		}
	case ActivityTypeCurrencyChange:
		details["new_currency"] = r.FormValue("currency_code")
	}

	if len(details) > 0 {
		detailsJSON, err := json.Marshal(details)
		if err == nil {
			activity.Details = string(detailsJSON)
		}
	}

	// Log the activity
	if err := LogActivity(activity); err != nil {
		m.log.Warnf("Failed to log activity: %v", err)
	}
}

// Import context keys from main package
// Use the context key types from the main package
type ctxKeySessionID = main.CtxKeySessionID
type ctxKeyRequestID = main.CtxKeyRequestID

type responseRecorder struct {
	w      http.ResponseWriter
	status int
}

func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.w.Write(b)
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.w.WriteHeader(status)
}

// currentCurrency gets the currency from the request context
func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie("currency")
	if c != nil {
		return c.Value
	}
	return "USD"
}