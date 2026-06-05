//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_CannedResponseCRUD(t *testing.T) {
	c := ticketAdminClient(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/canned-responses", map[string]any{
		"title":    fmt.Sprintf("e2e-canned-%d", suffix),
		"content":  "Thanks for contacting us, {{name}}.",
		"shortcut": fmt.Sprintf("ec%d", suffix%100000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	cannedID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/canned-responses/"+cannedID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/canned-responses", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/canned-responses/"+cannedID, map[string]any{
		"content": "Updated canned content",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/canned-responses/"+cannedID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
