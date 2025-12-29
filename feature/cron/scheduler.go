package cron

import (
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler manages cron jobs
type Scheduler struct {
	cron    *cron.Cron
	service *Service
}

// NewScheduler creates a new cron scheduler
func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		service: NewService(db),
	}
}

// Start starts all cron jobs
func (s *Scheduler) Start() error {
	// Mark overdue invoices - every hour
	_, err := s.cron.AddFunc("0 * * * *", func() {
		s.service.MarkOverdueInvoices()
	})
	if err != nil {
		return err
	}

	// Cleanup expired sessions - every day at 2 AM
	_, err = s.cron.AddFunc("0 2 * * *", func() {
		s.service.CleanupExpiredSessions()
	})
	if err != nil {
		return err
	}

	// Send payment reminders - every day at 9 AM
	_, err = s.cron.AddFunc("0 9 * * *", func() {
		s.service.SendPaymentReminders()
	})
	if err != nil {
		return err
	}

	// Update invoice statuses - every 6 hours
	_, err = s.cron.AddFunc("0 */6 * * *", func() {
		s.service.UpdateInvoiceStatuses()
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	return nil
}

// Stop stops all cron jobs
func (s *Scheduler) Stop() {
	s.cron.Stop()
}
