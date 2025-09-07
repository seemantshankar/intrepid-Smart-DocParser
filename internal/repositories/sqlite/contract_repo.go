package sqlite

import (
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repositories"
	"gorm.io/gorm"
)

type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository creates a new SQLite contract repository
func NewContractRepository(db *gorm.DB) repositories.ContractRepository {
	// Auto-migrate the schema
	err := db.AutoMigrate(&models.Contract{})
	if err != nil {
		panic("failed to migrate contract model: " + err.Error())
	}

	return &contractRepository{
		db: db,
	}
}

func (r *contractRepository) Create(c *models.Contract) error {
	return r.db.Create(c).Error
}

func (r *contractRepository) GetByID(id string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.First(&contract, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}
	return &contract, nil
}

func (r *contractRepository) Update(c *models.Contract) error {
	return r.db.Save(c).Error
}

func (r *contractRepository) Delete(id string) error {
	result := r.db.Delete(&models.Contract{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *contractRepository) List() ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.Find(&contracts).Error
	if err != nil {
		return nil, err
	}
	return contracts, nil
}
