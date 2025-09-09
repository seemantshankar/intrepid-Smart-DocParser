package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository represents a generic repository interface
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*T, error)
}

// ContractRepository defines the interface for contract operations
type ContractRepository interface {
	Repository[Contract]
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*Contract, error)
	GetByTitle(ctx context.Context, title string) (*Contract, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// MilestoneRepository defines the interface for milestone operations
type MilestoneRepository interface {
	Repository[Milestone]
	GetByContractID(ctx context.Context, contractID uuid.UUID, limit, offset int) ([]*Milestone, error)
	GetDueSoon(ctx context.Context, days int) ([]*Milestone, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// RiskAssessmentRepository defines the interface for risk assessment operations
type RiskAssessmentRepository interface {
	Repository[RiskAssessment]
	GetByContractID(ctx context.Context, contractID uuid.UUID, limit, offset int) ([]*RiskAssessment, error)
	GetByRiskLevel(ctx context.Context, level string, limit, offset int) ([]*RiskAssessment, error)
}

// KnowledgeEntryRepository defines the interface for knowledge entry operations
type KnowledgeEntryRepository interface {
	Repository[KnowledgeEntry]
	SearchByTags(ctx context.Context, tags []string, limit, offset int) ([]*KnowledgeEntry, error)
	SearchByContent(ctx context.Context, query string, limit, offset int) ([]*KnowledgeEntry, error)
}

// Entity models
type Contract struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	Status      string     `json:"status" gorm:"not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type Milestone struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID uuid.UUID  `json:"contract_id" gorm:"type:uuid;not null"`
	Name       string     `json:"name" gorm:"not null"`
	DueDate    *time.Time `json:"due_date"`
	Status     string     `json:"status" gorm:"not null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type RiskAssessment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID uuid.UUID `json:"contract_id" gorm:"type:uuid;not null"`
	RiskLevel  string    `json:"risk_level" gorm:"not null"`
	Description string   `json:"description"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
