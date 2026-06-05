//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_ExtendedActions(t *testing.T) {
	c := ticketAdminClient(t)

	resp, body, err := c.Do(http.MethodPost, "/api/v1/tickets", map[string]any{
		"subject": "E2E extended", "description": "status rate delete", "priority": "normal",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	id := e2e.ExtractID(t, body)

	var detail map[string]any
	e2e.ParseJSON(t, body, &detail)
	number := e2e.GetString(detail, "data", "number")
	if number == "" {
		number = e2e.GetString(detail, "data", "ticketNumber")
	}
	if number != "" {
		resp, body, err = c.Do(http.MethodGet, "/api/v1/tickets/number/"+number, nil)
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK)
	}

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/tickets/"+id+"/status", map[string]any{
		"status": "in_progress",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusBadRequest)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets/"+id+"/rate", map[string]any{
		"rating": 5, "comment": "great support",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusBadRequest)

	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	_, adminID, _, _ := e2e.AdminAuthContext(t, authURL, "admin@example.com", "AdminPass123!")
	resp, body, err = c.Do(http.MethodGet, "/api/v1/customers/"+adminID+"/tickets", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/tickets/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent, http.StatusBadRequest)
}
