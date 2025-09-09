package models

import (
	"time"

	"github.com/google/uuid"
)

// KnowledgeEntry represents a knowledge entry in the system.
type KnowledgeEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Industry  string    `gorm:"type:varchar(100);not null" json:"industry"`
	Source    string    `gorm:"type:varchar(500)" json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for GORM.
func (KnowledgeEntry) TableName() string {
	return "knowledge_entries"
}