//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_AssignTransferAndAgentTickets(t *testing.T) {
	c := ticketAdminClient(t)
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	agentUserID, agentEmail := registerTicketAgentUser(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/agents", map[string]any{
		"userId": agentUserID, "name": "Assign Agent", "email": agentEmail, "role": "agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	agentID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/departments", map[string]any{
		"name": fmt.Sprintf("assign-dept-%d", suffix), "description": "assign flow",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	deptID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets", map[string]any{
		"subject": fmt.Sprintf("assign-ticket-%d", suffix), "description": "assign test", "priority": "medium",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	ticketID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets/"+ticketID+"/assign", map[string]any{
		"assigneeId": agentUserID, "comment": "e2e assign",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/agents/"+agentUserID+"/tickets", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets/"+ticketID+"/transfer", map[string]any{
		"departmentId": deptID, "comment": "e2e transfer",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/tickets/"+ticketID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/agents/"+agentID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/departments/"+deptID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)

	_ = authURL
}
