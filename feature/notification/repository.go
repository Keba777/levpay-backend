package notification

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles notification-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new notification repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateNotification records a new notification
func (r *Repository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// GetUserNotifications retrieves all notifications for a user with pagination
func (r *Repository) GetUserNotifications(userID uuid.UUID, req models.ListedRequest) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read
func (r *Repository) MarkAsRead(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("read", true).Error
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *Repository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetNotificationByID retrieves a notification by ID
func (r *Repository) GetNotificationByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.Where("id = ?", id).First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}
