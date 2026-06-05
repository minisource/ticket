//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_BulkOperationsAndDepartmentAgents(t *testing.T) {
	c := ticketAdminClient(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/departments", map[string]any{
		"name": fmt.Sprintf("bulk-dept-%d", suffix), "description": "bulk ops dept",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	deptID := e2e.ExtractID(t, body)

	agentUserID, agentEmail := registerTicketAgentUser(t)
	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/agents", map[string]any{
		"userId": agentUserID, "name": "Bulk Dept Agent", "email": agentEmail, "role": "agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	agentID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/departments/"+deptID+"/agents", map[string]any{
		"agentId": agentUserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusBadRequest)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/departments/"+deptID+"/agents", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	catName := fmt.Sprintf("bulk-cat-%d", suffix)
	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/categories", map[string]any{
		"name": catName, "departmentId": deptID,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	catID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/categories", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/categories/"+catID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/categories/"+catID, map[string]any{
		"name": catName, "description": "updated via e2e",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	var ticketIDs []string
	for i := 0; i < 2; i++ {
		resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets", map[string]any{
			"subject": fmt.Sprintf("bulk-ext-%d-%d", suffix, i), "description": "bulk extended", "priority": "low",
		})
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
		ticketIDs = append(ticketIDs, e2e.ExtractID(t, body))
	}

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/tickets/bulk-status", map[string]any{
		"ticketIds": ticketIDs, "status": "open",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/tickets/bulk-priority", map[string]any{
		"ticketIds": ticketIDs, "priority": "high",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/tickets/bulk-transfer", map[string]any{
		"ticketIds": ticketIDs, "departmentId": deptID,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/tickets/bulk-delete", map[string]any{
		"ticketIds": ticketIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/departments/"+deptID+"/agents/"+agentUserID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent, http.StatusNotFound)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/categories/"+catID, nil)
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
}
