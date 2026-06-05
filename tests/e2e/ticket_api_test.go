//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_API(t *testing.T) {
	tenant := e2e.TenantHeader("default")
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	for k, v := range tenant {
		h[k] = v
	}

	c := e2e.NewClient(e2e.BaseURLFromEnv("TICKET_BASE_URL", "http://127.0.0.1:5011"), tenant)
	c.RequireUp(t, "/health")

	c.RunCases(t, []e2e.Case{
		{Name: "health", Method: http.MethodGet, Path: "/health", WantCode: []int{http.StatusOK}},
		{Name: "ready", Method: http.MethodGet, Path: "/ready", WantCode: []int{http.StatusOK}},
		{Name: "live", Method: http.MethodGet, Path: "/live", WantCode: []int{http.StatusOK}},
		{Name: "departments", Method: http.MethodGet, Path: "/api/v1/departments", WantCode: []int{http.StatusOK}},
		{Name: "categories", Method: http.MethodGet, Path: "/api/v1/categories", WantCode: []int{http.StatusOK}},
		{Name: "tickets_list", Method: http.MethodGet, Path: "/api/v1/tickets", WantCode: []int{http.StatusOK, http.StatusUnauthorized}},
		{Name: "tickets_stats", Method: http.MethodGet, Path: "/api/v1/tickets/stats", WantCode: []int{http.StatusOK, http.StatusUnauthorized}},
		{Name: "tickets_create", Method: http.MethodPost, Path: "/api/v1/tickets", Headers: h, Body: map[string]any{
			"subject": "E2E ticket", "description": "test", "priority": "normal",
		}, WantCode: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest}},
		{Name: "admin_agents", Method: http.MethodGet, Path: "/api/v1/admin/agents", Headers: h, WantCode: []int{http.StatusOK, http.StatusForbidden, http.StatusUnauthorized}},
		{Name: "admin_departments", Method: http.MethodGet, Path: "/api/v1/admin/departments", Headers: h, WantCode: []int{http.StatusOK, http.StatusForbidden, http.StatusUnauthorized}},
	})
}
