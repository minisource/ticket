//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func ticketAdminClient(t *testing.T) *e2e.Client {
	t.Helper()
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"
	c := e2e.NewClient(e2e.BaseURLFromEnv("TICKET_BASE_URL", "http://127.0.0.1:5011"), h)
	c.RequireUp(t, "/health")
	return c
}

func registerTicketAgentUser(t *testing.T) (userID, email string) {
	t.Helper()
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	auth := e2e.NewClient(authURL, nil)
	email = e2e.UniqueEmail("ticket-agent")
	password := "TestPass123"

	resp, body, err := auth.Do(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "firstName": "Ticket", "lastName": "Agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)

	token := e2e.LoginAuth(t, authURL, email, password)
	me := e2e.NewClient(authURL, e2e.Bearer(token))
	resp, body, err = me.Do(http.MethodGet, "/api/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	var parsed map[string]any
	e2e.ParseJSON(t, body, &parsed)
	userID = e2e.GetString(parsed, "data", "id")
	if userID == "" {
		userID = e2e.GetString(parsed, "id")
	}
	if userID == "" {
		t.Fatalf("no user id from /users/me: %s", string(body))
	}
	return userID, email
}

func TestTicket_AdminDepartmentAndCategoryCRUD(t *testing.T) {
	c := ticketAdminClient(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/departments", map[string]any{
		"name": fmt.Sprintf("e2e-dept-%d", suffix), "description": "e2e department",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	deptID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/departments/"+deptID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/departments/"+deptID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/departments/"+deptID, map[string]any{
		"description": "updated dept",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/categories", map[string]any{
		"name": fmt.Sprintf("e2e-cat-%d", suffix), "departmentId": deptID,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	catID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/categories/"+catID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/categories/"+catID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/departments/"+deptID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}

func TestTicket_AdminAgentAndDashboard(t *testing.T) {
	c := ticketAdminClient(t)
	userID, email := registerTicketAgentUser(t)

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/agents", map[string]any{
		"userId": userID, "name": "E2E Agent", "email": email, "role": "agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	agentID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/agents/"+agentID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/agents/"+agentID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
