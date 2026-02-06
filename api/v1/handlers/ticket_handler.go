package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/ticket/internal/models"
	"github.com/minisource/ticket/internal/usecase"
)

// TicketHandler handles ticket HTTP requests
type TicketHandler struct {
	ticketUsecase *usecase.TicketUsecase
	translator    *i18n.Translator
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketUsecase *usecase.TicketUsecase) *TicketHandler {
	return &TicketHandler{
		ticketUsecase: ticketUsecase,
		translator:    i18n.GetTranslator(),
	}
}

// CreateTicket creates a new ticket
// @Summary Create a new ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param ticket body models.CreateTicketRequest true "Ticket data"
// @Success 201 {object} Response{data=models.Ticket}
// @Failure 400 {object} Response
// @Router /api/v1/tickets [post]
func (h *TicketHandler) CreateTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")
	userEmail := c.Get("X-User-Email")
	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	if userID == "" {
		return response.BadRequest(c, "MISSING_USER", h.translator.Translate(ctx, "error.missing_user_id", nil))
	}

	var req models.CreateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	// Set tenant from header if not in request
	if req.TenantID == "" {
		req.TenantID = c.Get("X-Tenant-ID")
	}

	ticket, err := h.ticketUsecase.CreateTicket(ctx, req, userID, userName, userEmail, ip, userAgent)
	if err != nil {
		return response.BadRequest(c, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, ticket)
}

// GetTicket gets a ticket by ID
// @Summary Get a ticket by ID
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Ticket ID"
// @Success 200 {object} Response{data=models.Ticket}
// @Failure 404 {object} Response
// @Router /api/v1/tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	ticket, err := h.ticketUsecase.GetTicket(ctx, id)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "ticket.not_found", nil))
	}

	return response.OK(c, ticket)
}

// GetTicketByNumber gets a ticket by number
// @Summary Get a ticket by number
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param number path string true "Ticket Number"
// @Success 200 {object} Response{data=models.Ticket}
// @Failure 404 {object} Response
// @Router /api/v1/tickets/number/{number} [get]
func (h *TicketHandler) GetTicketByNumber(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	number := c.Params("number")

	ticket, err := h.ticketUsecase.GetTicketByNumber(ctx, tenantID, number)
	if err != nil {
		return response.NotFound(c, h.translator.Translate(ctx, "ticket.not_found", nil))
	}

	return response.OK(c, ticket)
}

// ListTickets lists tickets
// @Summary List tickets
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param status query string false "Status filter (comma-separated)"
// @Param priority query string false "Priority filter (comma-separated)"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.Ticket}
// @Router /api/v1/tickets [get]
func (h *TicketHandler) ListTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")

	filter := models.TicketFilter{
		TenantID: tenantID,
		Page:     1,
		PerPage:  20,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			filter.PerPage = pp
		}
	}
	if status := c.Query("status"); status != "" {
		statuses := strings.Split(status, ",")
		for _, s := range statuses {
			filter.Status = append(filter.Status, models.TicketStatus(strings.TrimSpace(s)))
		}
	}
	if priority := c.Query("priority"); priority != "" {
		priorities := strings.Split(priority, ",")
		for _, p := range priorities {
			filter.Priority = append(filter.Priority, models.TicketPriority(strings.TrimSpace(p)))
		}
	}
	if departmentID := c.Query("department_id"); departmentID != "" {
		filter.DepartmentID = departmentID
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		filter.CategoryID = categoryID
	}
	if assignedToID := c.Query("assigned_to"); assignedToID != "" {
		filter.AssignedToID = assignedToID
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	tickets, total, err := h.ticketUsecase.ListTickets(ctx, filter)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, tickets, &response.Pagination{
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		Total:      total,
		TotalPages: int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage)),
	})
}

// GetMyTickets gets tickets for the current user
// @Summary Get my tickets
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.Ticket}
// @Router /api/v1/tickets/my [get]
func (h *TicketHandler) GetMyTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")

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

	tickets, total, err := h.ticketUsecase.GetCustomerTickets(ctx, tenantID, userID, page, perPage)
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

// UpdateTicket updates a ticket
// @Summary Update a ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param id path string true "Ticket ID"
// @Param ticket body models.UpdateTicketRequest true "Update data"
// @Success 200 {object} Response{data=models.Ticket}
// @Failure 400 {object} Response
// @Router /api/v1/tickets/{id} [patch]
func (h *TicketHandler) UpdateTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.UpdateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	// isAgent = false for customer routes
	ticket, err := h.ticketUsecase.UpdateTicket(ctx, id, req, userID, userName, false)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// AddReply adds a reply to a ticket
// @Summary Add a reply to a ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param id path string true "Ticket ID"
// @Param message body models.CreateMessageRequest true "Message data"
// @Success 201 {object} Response{data=models.TicketMessage}
// @Failure 400 {object} Response
// @Router /api/v1/tickets/{id}/reply [post]
func (h *TicketHandler) AddReply(c *fiber.Ctx) error {
	ctx := c.Context()
	ticketID := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")
	userEmail := c.Get("X-User-Email")
	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	var req models.CreateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	message, err := h.ticketUsecase.AddReply(ctx, ticketID, req, userID, userName, userEmail, models.SenderCustomer, ip, userAgent)
	if err != nil {
		return response.BadRequest(c, "ADD_REPLY_FAILED", err.Error())
	}

	return response.Created(c, message)
}

// GetTicketMessages gets messages for a ticket
// @Summary Get ticket messages
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Ticket ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.TicketMessage}
// @Router /api/v1/tickets/{id}/messages [get]
func (h *TicketHandler) GetTicketMessages(c *fiber.Ctx) error {
	ctx := c.Context()
	ticketID := c.Params("id")

	page := 1
	perPage := 50
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

	// includePrivate = false for customers
	messages, total, err := h.ticketUsecase.GetTicketMessages(ctx, ticketID, false, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, messages, &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// GetTicketHistory gets history for a ticket
// @Summary Get ticket history
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Ticket ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.TicketHistory}
// @Router /api/v1/tickets/{id}/history [get]
func (h *TicketHandler) GetTicketHistory(c *fiber.Ctx) error {
	ctx := c.Context()
	ticketID := c.Params("id")

	page := 1
	perPage := 50
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

	history, total, err := h.ticketUsecase.GetTicketHistory(ctx, ticketID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, history, &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// ChangeStatus changes ticket status
// @Summary Change ticket status
// @Tags Tickets
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param id path string true "Ticket ID"
// @Param status body models.ChangeStatusRequest true "Status data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/tickets/{id}/status [patch]
func (h *TicketHandler) ChangeStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.ChangeStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	// isAgent = false for customer routes
	ticket, err := h.ticketUsecase.ChangeStatus(ctx, id, req, userID, userName, false)
	if err != nil {
		return response.BadRequest(c, "STATUS_CHANGE_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// RateTicket rates a ticket
// @Summary Rate a ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param id path string true "Ticket ID"
// @Param rating body models.RateTicketRequest true "Rating data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/tickets/{id}/rate [post]
func (h *TicketHandler) RateTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")

	var req models.RateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	ticket, err := h.ticketUsecase.RateTicket(ctx, id, req, userID)
	if err != nil {
		return response.BadRequest(c, "RATE_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// ===== Agent-specific endpoints =====

// AgentAddReply adds a reply as an agent
// @Summary Add a reply as an agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param message body models.CreateMessageRequest true "Message data"
// @Success 201 {object} Response{data=models.TicketMessage}
// @Router /api/v1/agent/tickets/{id}/reply [post]
func (h *TicketHandler) AgentAddReply(c *fiber.Ctx) error {
	ctx := c.Context()
	ticketID := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")
	userEmail := c.Get("X-User-Email")
	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	var req models.CreateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	message, err := h.ticketUsecase.AddReply(ctx, ticketID, req, userID, userName, userEmail, models.SenderAgent, ip, userAgent)
	if err != nil {
		return response.BadRequest(c, "ADD_REPLY_FAILED", err.Error())
	}

	return response.Created(c, message)
}

// AgentChangeStatus changes ticket status as an agent
// @Summary Change ticket status as agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param status body models.ChangeStatusRequest true "Status data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/agent/tickets/{id}/status [patch]
func (h *TicketHandler) AgentChangeStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.ChangeStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	// isAgent = true for agent routes
	ticket, err := h.ticketUsecase.ChangeStatus(ctx, id, req, userID, userName, true)
	if err != nil {
		return response.BadRequest(c, "STATUS_CHANGE_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// AgentUpdateTicket updates a ticket as an agent
// @Summary Update a ticket as agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param ticket body models.UpdateTicketRequest true "Update data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/agent/tickets/{id} [patch]
func (h *TicketHandler) AgentUpdateTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.UpdateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	// isAgent = true for agent routes
	ticket, err := h.ticketUsecase.UpdateTicket(ctx, id, req, userID, userName, true)
	if err != nil {
		return response.BadRequest(c, "UPDATE_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// AgentAssignTicket assigns a ticket to an agent
// @Summary Assign ticket to agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param assign body models.AssignTicketRequest true "Assign data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/agent/tickets/{id}/assign [post]
func (h *TicketHandler) AgentAssignTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.AssignTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	ticket, err := h.ticketUsecase.AssignTicket(ctx, id, req, userID, userName)
	if err != nil {
		return response.BadRequest(c, "ASSIGN_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// AgentTransferTicket transfers a ticket to a department
// @Summary Transfer ticket to department
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param transfer body models.TransferTicketRequest true "Transfer data"
// @Success 200 {object} Response{data=models.Ticket}
// @Router /api/v1/agent/tickets/{id}/transfer [post]
func (h *TicketHandler) AgentTransferTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")

	var req models.TransferTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", h.translator.Translate(ctx, "error.invalid_request_body", nil))
	}

	ticket, err := h.ticketUsecase.TransferTicket(ctx, id, req, userID, userName)
	if err != nil {
		return response.BadRequest(c, "TRANSFER_FAILED", err.Error())
	}

	return response.OK(c, ticket)
}

// AgentGetMessages gets messages including private notes
// @Summary Get ticket messages as agent
// @Tags Agent
// @Produce json
// @Param id path string true "Ticket ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.TicketMessage}
// @Router /api/v1/agent/tickets/{id}/messages [get]
func (h *TicketHandler) AgentGetMessages(c *fiber.Ctx) error {
	ctx := c.Context()
	ticketID := c.Params("id")

	page := 1
	perPage := 50
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

	// includePrivate = true for agents
	messages, total, err := h.ticketUsecase.GetTicketMessages(ctx, ticketID, true, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OKWithPagination(c, messages, &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// AgentGetMyTickets gets tickets assigned to the current agent
// @Summary Get my assigned tickets
// @Tags Agent
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.Ticket}
// @Router /api/v1/agent/tickets/my [get]
func (h *TicketHandler) AgentGetMyTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Get("X-User-ID")

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

	tickets, total, err := h.ticketUsecase.GetAgentTickets(ctx, tenantID, userID, page, perPage)
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

// GetStats gets ticket statistics
// @Summary Get ticket statistics
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} Response{data=models.TicketStats}
// @Router /api/v1/tickets/stats [get]
func (h *TicketHandler) GetStats(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")

	stats, err := h.ticketUsecase.GetStats(ctx, tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, stats)
}

// DeleteTicket deletes a ticket
// @Summary Delete a ticket
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-User-ID header string true "User ID"
// @Param id path string true "Ticket ID"
// @Success 200 {object} Response
// @Router /api/v1/tickets/{id} [delete]
func (h *TicketHandler) DeleteTicket(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	userID := c.Get("X-User-ID")

	if err := h.ticketUsecase.DeleteTicket(ctx, id, userID); err != nil {
		return response.BadRequest(c, "DELETE_FAILED", err.Error())
	}

	return response.OK(c, map[string]string{"message": h.translator.Translate(ctx, "ticket.deleted", nil)})
}

// GetCustomerTickets gets tickets for a specific customer (admin view)
// @Summary Get tickets for a customer
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param customer_id path string true "Customer ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.Ticket}
// @Router /api/v1/customers/{customer_id}/tickets [get]
func (h *TicketHandler) GetCustomerTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	customerID := c.Params("customer_id")

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

	tickets, total, err := h.ticketUsecase.GetCustomerTickets(ctx, tenantID, customerID, page, perPage)
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

// GetAgentTickets gets tickets assigned to a specific agent (admin view)
// @Summary Get tickets for an agent
// @Tags Tickets
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param agent_id path string true "Agent ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} Response{data=[]models.Ticket}
// @Router /api/v1/agents/{agent_id}/tickets [get]
func (h *TicketHandler) GetAgentTickets(c *fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Get("X-Tenant-ID")
	agentID := c.Params("agent_id")

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

	tickets, total, err := h.ticketUsecase.GetAgentTickets(ctx, tenantID, agentID, page, perPage)
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

// AssignTicket assigns a ticket to an agent (used by router)
// Redirects to AgentAssignTicket
func (h *TicketHandler) AssignTicket(c *fiber.Ctx) error {
	return h.AgentAssignTicket(c)
}

// TransferTicket transfers a ticket to a department (used by router)
// Redirects to AgentTransferTicket
func (h *TicketHandler) TransferTicket(c *fiber.Ctx) error {
	return h.AgentTransferTicket(c)
}
