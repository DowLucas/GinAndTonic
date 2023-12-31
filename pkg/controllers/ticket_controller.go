package controllers

import (
	"net/http"
	"strconv"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"github.com/DowLucas/gin-ticket-release/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketController struct {
	DB      *gorm.DB
	Service *services.TicketService
}

func NewTicketController(db *gorm.DB) *TicketController {
	service := services.NewTicketService(db)
	return &TicketController{DB: db, Service: service}
}

func (tc *TicketController) ListTickets(c *gin.Context) {

	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tickets, err := tc.Service.GetAllTicketsToEvent(eventID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	println("Tickets: ", tickets)

	c.JSON(http.StatusOK, tickets)
}

func (tc *TicketController) GetTicket(c *gin.Context) {
	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketIDstring := c.Param("ticketID")
	ticketID, err := strconv.Atoi(ticketIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := tc.Service.GetTicketToEvent(eventID, ticketID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (tc *TicketController) EditTicket(c *gin.Context) {
	var ticket models.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventIDstring := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketIDstring := c.Param("ticketID")
	ticketID, err := strconv.Atoi(ticketIDstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err = tc.Service.EditTicket(eventID, ticketID, ticket)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (tc *TicketController) UsersList(c *gin.Context) {
	// Find all ticket requests for the user

	UGKthId, _ := c.Get("ugkthid")

	ticketRequests, err := tc.Service.GetTicketForUser(UGKthId.(string))

	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": ticketRequests})
}
