package postgres

import (
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repositories"
	"gorm.io/gorm"
)

type contractRepo struct {
	db *gorm.DB
}

func NewContractRepository(db *gorm.DB) repositories.ContractRepository {
	return &contractRepo{db: db}
}

func (r *contractRepo) Create(c *models.Contract) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(c).Error
	})
}

func (r *contractRepo) GetByID(id string) (*models.Contract, error) {
	var c models.Contract
	err := r.db.Preload("Milestones").Preload("Risks").First(&c, "id = ?", id).Error
	return &c, err
}

func (r *contractRepo) Update(c *models.Contract) error {
	return r.db.Save(c).Error
}

func (r *contractRepo) Delete(id string) error {
	return r.db.Delete(&models.Contract{}, "id = ?", id).Error
}

func (r *contractRepo) List() ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.Find(&contracts).Error
	return contracts, err
}
