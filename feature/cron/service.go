package cron

import (
	"time"

	"github.com/Keba777/levpay-backend/feature/billing"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"gorm.io/gorm"
)

// Service handles cron job operations
type Service struct {
	db          *gorm.DB
	billingRepo *billing.Repository
	logger      *utils.Logger
}

// NewService creates a new cron service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:          db,
		billingRepo: billing.NewRepository(db),
		logger:      utils.GetLogger("cron"),
	}
}

// MarkOverdueInvoices marks invoices as overdue if past due date
func (s *Service) MarkOverdueInvoices() error {
	s.logger.Info("Running: Mark overdue invoices")

	overdueInvoices, err := s.billingRepo.GetOverdueInvoices()
	if err != nil {
		s.logger.ErrorWithErr("Failed to get overdue invoices", err)
		return err
	}

	count := 0
	for _, invoice := range overdueInvoices {
		if invoice.Status != models.InvoiceStatusPaid && invoice.Status != models.InvoiceStatusCancelled {
			if err := s.billingRepo.UpdateInvoiceStatus(invoice.ID, models.InvoiceStatusOverdue); err != nil {
				s.logger.ErrorWithErr("Failed to mark invoice as overdue", err)
				continue
			}
			count++
		}
	}

	s.logger.Info("Marked invoices as overdue", utils.Field{Key: "count", Value: count})
	return nil
}

// CleanupExpiredSessions removes old JWT sessions
func (s *Service) CleanupExpiredSessions() error {
	s.logger.Info("Running: Cleanup expired sessions")

	// Delete sessions older than 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	result := s.db.Where("updated_at < ?", thirtyDaysAgo).Delete(&models.Session{})
	if result.Error != nil {
		s.logger.ErrorWithErr("Failed to cleanup sessions", result.Error)
		return result.Error
	}

	s.logger.Info("Cleaned up expired sessions", utils.Field{Key: "count", Value: result.RowsAffected})
	return nil
}

// SendPaymentReminders sends reminders for invoices due soon
func (s *Service) SendPaymentReminders() error {
	s.logger.Info("Running: Send payment reminders")

	// Find invoices due in 3 days
	threeDaysFromNow := time.Now().AddDate(0, 0, 3)

	var upcomingInvoices []models.Invoice
	err := s.db.Where("status = ? AND due_date BETWEEN ? AND ?",
		models.InvoiceStatusSent,
		time.Now(),
		threeDaysFromNow).
		Find(&upcomingInvoices).Error

	if err != nil {
		s.logger.ErrorWithErr("Failed to get upcoming invoices", err)
		return err
	}

	// TODO: Send notifications via RabbitMQ
	s.logger.Info("Payment reminders to send", utils.Field{Key: "count", Value: len(upcomingInvoices)})

	return nil
}

// UpdateInvoiceStatuses performs general invoice status maintenance
func (s *Service) UpdateInvoiceStatuses() error {
	s.logger.Info("Running: Update invoice statuses")

	// Mark draft invoices as sent if they have a due date
	result := s.db.Model(&models.Invoice{}).
		Where("status = ? AND due_date IS NOT NULL", models.InvoiceStatusDraft).
		Update("status", models.InvoiceStatusSent)

	if result.Error != nil {
		s.logger.ErrorWithErr("Failed to update invoice statuses", result.Error)
		return result.Error
	}

	s.logger.Info("Updated invoice statuses", utils.Field{Key: "count", Value: result.RowsAffected})
	return nil
}
