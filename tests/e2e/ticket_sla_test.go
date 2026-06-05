//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestTicket_SLAPolicyCRUD(t *testing.T) {
	c := ticketAdminClient(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/sla-policies", map[string]any{
		"name":        fmt.Sprintf("e2e-sla-%d", suffix),
		"description": "e2e SLA policy",
		"priorities": []map[string]any{
			{
				"priority":            "medium",
				"firstResponseMins":   60,
				"resolutionMins":      240,
				"nextResponseMins":    30,
				"escalationEnabled":   false,
				"escalationAfterMins": 0,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks ticket admin permissions")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	policyID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/sla-policies/"+policyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/sla-policies", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPatch, "/api/v1/admin/sla-policies/"+policyID, map[string]any{
		"description": "updated SLA",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/sla-policies/"+policyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
