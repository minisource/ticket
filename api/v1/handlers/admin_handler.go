package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/ticket/internal/models"
	"github.com/minisource/ticket/internal/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	adminUsecase      *usecase.AdminUsecase
	departmentUsecase *usecase.DepartmentUsecase
	categoryUsecase   *usecase.CategoryUsecase
	translator        *i18n.Translator
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	adminUsecase *usecase.AdminUsecase,
	departmentUsecase *usecase.DepartmentUsecase,
	categoryUsecase *usecase.CategoryUsecase,
) *AdminHandler {
	return &AdminHandler{
		adminUsecase:      adminUsecase,
		departmentUsecase: departmentUsecase,
		categoryUsecase:   categoryUsecase,
		translator:        i18n.GetTranslator(),
	}
}

// ===== Agent Management =====

// CreateAgent creates a new agent
func (h *AdminHandler) CreateAgent(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var req models.CreateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	agent, err := h.adminUsecase.CreateAgent(ctx, tenantID, req)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, agent)
}

// GetAgent gets an agent by ID
func (h *AdminHandler) GetAgent(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	agent, err := h.adminUsecase.GetAgent(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "agent.not_found", nil))
	}

	return response.OK(c, agent)
}

// UpdateAgent updates an agent
func (h *AdminHandler) UpdateAgent(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req models.UpdateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	agent, err := h.adminUsecase.UpdateAgent(ctx, id, req)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, agent)
}

// DeleteAgent deletes an agent
func (h *AdminHandler) DeleteAgent(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	err := h.adminUsecase.DeleteAgent(ctx, id)
	if err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "agent.deleted", nil)})
}

// ListAgents lists agents
func (h *AdminHandler) ListAgents(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	activeOnly := c.Query("active_only") == "true"

	agents, err := h.adminUsecase.ListAgents(ctx, tenantID, activeOnly)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, agents)
}

// UpdateAgentStatus updates an agent's status
func (h *AdminHandler) UpdateAgentStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Params("id")

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	err := h.adminUsecase.UpdateAgentStatus(ctx, tenantID, userID, models.AgentStatus(req.Status))
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "agent.status_updated", nil)})
}

// ===== Department Management =====

// CreateDepartment creates a new department
func (h *AdminHandler) CreateDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var req models.CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	department, err := h.departmentUsecase.CreateDepartment(ctx, tenantID, req)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, department)
}

// GetDepartment gets a department by ID
func (h *AdminHandler) GetDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	department, err := h.departmentUsecase.GetDepartment(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "department.not_found", nil))
	}

	return response.OK(c, department)
}

// UpdateDepartment updates a department
func (h *AdminHandler) UpdateDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req models.UpdateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	department, err := h.departmentUsecase.UpdateDepartment(ctx, id, req)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, department)
}

// DeleteDepartment deletes a department
func (h *AdminHandler) DeleteDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	err := h.departmentUsecase.DeleteDepartment(ctx, id)
	if err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "department.deleted", nil)})
}

// ListDepartments lists departments
func (h *AdminHandler) ListDepartments(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	activeOnly := c.Query("active_only") == "true"

	departments, err := h.departmentUsecase.ListDepartments(ctx, tenantID, activeOnly)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, departments)
}

// AddAgentToDepartment adds an agent to a department
func (h *AdminHandler) AddAgentToDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req struct {
		AgentID string `json:"agentId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	err := h.departmentUsecase.AddAgentToDepartment(ctx, id, req.AgentID)
	if err != nil {
		return response.BadRequest(c, "ADD_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "department.agent_added", nil)})
}

// RemoveAgentFromDepartment removes an agent from a department
func (h *AdminHandler) RemoveAgentFromDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	agentID := c.Params("agentId")

	err := h.departmentUsecase.RemoveAgentFromDepartment(ctx, id, agentID)
	if err != nil {
		return response.BadRequest(c, "REMOVE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "department.agent_removed", nil)})
}

// GetDepartmentAgents gets all agents in a department
func (h *AdminHandler) GetDepartmentAgents(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	id := c.Params("id")

	agents, err := h.departmentUsecase.GetDepartmentAgents(ctx, tenantID, id)
	if err != nil {
		return response.BadRequest(c, "GET_AGENTS_FAILED", err.Error())
	}

	return response.OK(c, agents)
}

// ===== Category Management =====

// CreateCategory creates a new category
func (h *AdminHandler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	category, err := h.categoryUsecase.CreateCategory(ctx, tenantID, req)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, category)
}

// GetCategory gets a category by ID
func (h *AdminHandler) GetCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	category, err := h.categoryUsecase.GetCategory(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "category.not_found", nil))
	}

	return response.OK(c, category)
}

// UpdateCategory updates a category
func (h *AdminHandler) UpdateCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req models.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	category, err := h.categoryUsecase.UpdateCategory(ctx, id, req)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, category)
}

// DeleteCategory deletes a category
func (h *AdminHandler) DeleteCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	err := h.categoryUsecase.DeleteCategory(ctx, id)
	if err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "category.deleted", nil)})
}

// ListCategories lists categories
func (h *AdminHandler) ListCategories(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	publicOnly := c.Query("public_only") == "true"

	categories, err := h.categoryUsecase.ListCategories(ctx, tenantID, publicOnly)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, categories)
}

// ===== SLA Policy Management =====

// CreateSLAPolicy creates a new SLA policy
func (h *AdminHandler) CreateSLAPolicy(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var req models.CreateSLAPolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	policy, err := h.adminUsecase.CreateSLAPolicy(ctx, tenantID, req)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, policy)
}

// GetSLAPolicy gets an SLA policy by ID
func (h *AdminHandler) GetSLAPolicy(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	policy, err := h.adminUsecase.GetSLAPolicy(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "sla.not_found", nil))
	}

	return response.OK(c, policy)
}

// UpdateSLAPolicy updates an SLA policy
func (h *AdminHandler) UpdateSLAPolicy(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req models.UpdateSLAPolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	policy, err := h.adminUsecase.UpdateSLAPolicy(ctx, id, req)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, policy)
}

// DeleteSLAPolicy deletes an SLA policy
func (h *AdminHandler) DeleteSLAPolicy(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	err := h.adminUsecase.DeleteSLAPolicy(ctx, id)
	if err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "sla.deleted", nil)})
}

// ListSLAPolicies lists SLA policies
func (h *AdminHandler) ListSLAPolicies(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	policies, err := h.adminUsecase.ListSLAPolicies(ctx, tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, policies)
}

// ===== Canned Response Management =====

// CreateCannedResponse creates a new canned response
func (h *AdminHandler) CreateCannedResponse(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var req models.CreateCannedResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	cannedResp, err := h.adminUsecase.CreateCannedResponse(ctx, tenantID, userID, req)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, cannedResp)
}

// GetCannedResponse gets a canned response by ID
func (h *AdminHandler) GetCannedResponse(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	cannedResp, err := h.adminUsecase.GetCannedResponse(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "canned.not_found", nil))
	}

	return response.OK(c, cannedResp)
}

// UpdateCannedResponse updates a canned response
func (h *AdminHandler) UpdateCannedResponse(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	var req models.UpdateCannedResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	cannedResp, err := h.adminUsecase.UpdateCannedResponse(ctx, id, req)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, cannedResp)
}

// DeleteCannedResponse deletes a canned response
func (h *AdminHandler) DeleteCannedResponse(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	err := h.adminUsecase.DeleteCannedResponse(ctx, id)
	if err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "canned.deleted", nil)})
}

// ListCannedResponses lists canned responses
func (h *AdminHandler) ListCannedResponses(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	var departmentID *primitive.ObjectID
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		if deptID, err := primitive.ObjectIDFromHex(deptIDStr); err == nil {
			departmentID = &deptID
		}
	}

	globalOnly := c.Query("global_only") == "true"

	responses, err := h.adminUsecase.ListCannedResponses(ctx, tenantID, departmentID, globalOnly)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, responses)
}

// ===== Bulk Operations =====

// BulkAssignTickets assigns multiple tickets to an agent
func (h *AdminHandler) BulkAssignTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req struct {
		TicketIDs []string `json:"ticketIds"`
		AgentID   string   `json:"agentId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	count, err := h.adminUsecase.BulkAssignTickets(ctx, tenantID, req.TicketIDs, req.AgentID, userID, userName)
	if err != nil {
		return response.BadRequest(c, "BULK_ASSIGN_FAILED", err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"message":      h.translator.Translate(ctx, "ticket.bulk_assigned", nil),
		"successCount": count,
	})
}

// BulkChangeStatus changes status of multiple tickets
func (h *AdminHandler) BulkChangeStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req struct {
		TicketIDs []string `json:"ticketIds"`
		Status    string   `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	count, err := h.adminUsecase.BulkChangeStatus(ctx, tenantID, req.TicketIDs, models.TicketStatus(req.Status), userID, userName)
	if err != nil {
		return response.BadRequest(c, "BULK_STATUS_FAILED", err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"message":      h.translator.Translate(ctx, "ticket.bulk_status_changed", nil),
		"successCount": count,
	})
}

// BulkChangePriority changes priority of multiple tickets
func (h *AdminHandler) BulkChangePriority(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req struct {
		TicketIDs []string `json:"ticketIds"`
		Priority  string   `json:"priority"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	count, err := h.adminUsecase.BulkChangePriority(ctx, tenantID, req.TicketIDs, models.TicketPriority(req.Priority), userID, userName)
	if err != nil {
		return response.BadRequest(c, "BULK_PRIORITY_FAILED", err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"message":      h.translator.Translate(ctx, "ticket.bulk_priority_changed", nil),
		"successCount": count,
	})
}

// BulkTransferDepartment transfers multiple tickets to a department
func (h *AdminHandler) BulkTransferDepartment(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req struct {
		TicketIDs    []string `json:"ticketIds"`
		DepartmentID string   `json:"departmentId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	count, err := h.adminUsecase.BulkTransferDepartment(ctx, tenantID, req.TicketIDs, req.DepartmentID, userID, userName)
	if err != nil {
		return response.BadRequest(c, "BULK_TRANSFER_FAILED", err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"message":      h.translator.Translate(ctx, "ticket.bulk_transferred", nil),
		"successCount": count,
	})
}

// BulkDeleteTickets deletes multiple tickets
func (h *AdminHandler) BulkDeleteTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")

	var req struct {
		TicketIDs []string `json:"ticketIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	count, err := h.adminUsecase.BulkDeleteTickets(ctx, tenantID, req.TicketIDs)
	if err != nil {
		return response.BadRequest(c, "BULK_DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"message":      h.translator.Translate(ctx, "ticket.bulk_deleted", nil),
		"successCount": count,
	})
}

// ===== Dashboard & Statistics =====

// GetDashboardStats gets dashboard statistics
func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	stats, err := h.adminUsecase.GetDashboardStats(ctx, tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, stats)
}

// GetAgentStats gets agent statistics
func (h *AdminHandler) GetAgentStats(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	agentID := c.Params("id")

	agent, err := h.adminUsecase.GetAgentStats(ctx, tenantID, agentID)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "agent.not_found", nil))
	}

	return response.OK(c, agent)
}

// GetDepartmentStats gets department statistics
func (h *AdminHandler) GetDepartmentStats(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	dept, err := h.adminUsecase.GetDepartmentStats(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "department.not_found", nil))
	}

	return response.OK(c, dept)
}

// GetSLABreachedTickets gets tickets with breached SLA
func (h *AdminHandler) GetSLABreachedTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if pVal, err := strconv.Atoi(p); err == nil {
			page = pVal
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if ppVal, err := strconv.Atoi(pp); err == nil {
			perPage = ppVal
		}
	}

	tickets, total, err := h.adminUsecase.GetSLABreachedTickets(ctx, tenantID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, tickets, &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// GetTicketsDueSoon gets tickets due soon
func (h *AdminHandler) GetTicketsDueSoon(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	hours := 4
	if h := c.Query("hours"); h != "" {
		if hVal, err := strconv.Atoi(h); err == nil {
			hours = hVal
		}
	}

	tickets, err := h.adminUsecase.GetTicketsDueSoon(ctx, tenantID, hours)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, tickets)
}

// GetUnassignedTickets gets unassigned tickets
func (h *AdminHandler) GetUnassignedTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	if tenantID == "" {
		return response.BadRequest(c, "TENANT_REQUIRED", h.translator.Translate(ctx, "error.tenant_required", nil))
	}

	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if pVal, err := strconv.Atoi(p); err == nil {
			page = pVal
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if ppVal, err := strconv.Atoi(pp); err == nil {
			perPage = ppVal
		}
	}

	tickets, total, err := h.adminUsecase.GetUnassignedTickets(ctx, tenantID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, tickets, &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}
