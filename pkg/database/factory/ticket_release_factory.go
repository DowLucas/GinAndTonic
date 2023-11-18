package factory

import (
	"github.com/DowLucas/gin-ticket-release/pkg/models"
)

func NewTicketRelease(eventID int, open, close uint, hasAllocatedTickets bool, ticketReleaseMethodDetailID uint) *models.TicketRelease {
	return &models.TicketRelease{
		EventID:                     eventID,
		Open:                        open,
		Close:                       close,
		HasAllocatedTickets:         hasAllocatedTickets,
		TicketReleaseMethodDetailID: ticketReleaseMethodDetailID,
	}
}
