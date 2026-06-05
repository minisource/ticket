//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_CRUDFlow(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"

	c := e2e.NewClient(e2e.BaseURLFromEnv("TICKET_BASE_URL", "http://127.0.0.1:5011"), h)
	c.RequireUp(t, "/health")

	resp, body, err := c.Do(http.MethodPost, "/api/v1/tickets", map[string]any{
		"subject":     "E2E CRUD ticket",
		"description": "full flow test",
		"priority":    "normal",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)

	var created map[string]any
	e2e.ParseJSON(t, body, &created)
	id := e2e.GetString(created, "data", "id")
	if id == "" {
		id = e2e.GetString(created, "data", "_id")
	}
	if id == "" {
		t.Fatalf("no ticket id in create response: %s", string(body))
	}

	resp, body, err = c.Do(http.MethodGet, "/api/v1/tickets/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/tickets/"+id, map[string]any{
		"subject": "E2E CRUD ticket updated",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/tickets/"+id+"/messages", map[string]any{
		"content": "E2E reply message",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/tickets/"+id+"/messages", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/tickets/"+id+"/history", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)
}
