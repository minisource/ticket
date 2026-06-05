//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_BulkAssignAndUpdateAgent(t *testing.T) {
	c := ticketAdminClient(t)
	agentUserID, agentEmail := registerTicketAgentUser(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/agents", map[string]any{
		"userId": agentUserID, "name": "Bulk Agent", "email": agentEmail, "role": "agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	agentID := e2e.ExtractID(t, body)

	var ticketIDs []string
	for i := 0; i < 2; i++ {
		resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets", map[string]any{
			"subject": fmt.Sprintf("bulk-ticket-%d-%d", suffix, i), "description": "bulk assign", "priority": "medium",
		})
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
		ticketIDs = append(ticketIDs, e2e.ExtractID(t, body))
	}

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/tickets/bulk-assign", map[string]any{
		"ticketIds": ticketIDs, "agentId": agentUserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/agents/"+agentID, map[string]any{
		"name": "Bulk Agent Updated",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/agents/"+agentID+"/status", map[string]any{
		"status": "available",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	for _, tid := range ticketIDs {
		resp, body, err = c.Do(http.MethodDelete, "/api/v1/tickets/"+tid, nil)
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
	}

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/agents/"+agentID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
