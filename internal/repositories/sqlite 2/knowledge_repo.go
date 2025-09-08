package sqlite

import (
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repositories"
	"gorm.io/gorm"
)

// knowledgeRepo implements the repositories.KnowledgeEntryRepository interface for SQLite.
type knowledgeRepo struct {
	db *gorm.DB
}

// NewKnowledgeEntryRepository creates a new knowledge entry repository.
func NewKnowledgeEntryRepository(db *gorm.DB) repositories.KnowledgeEntryRepository {
	return &knowledgeRepo{db: db}
}

func (r *knowledgeRepo) Create(k *models.KnowledgeEntry) error {
	return r.db.Create(k).Error
}

func (r *knowledgeRepo) GetByID(id string) (*models.KnowledgeEntry, error) {
	var k models.KnowledgeEntry
	if err := r.db.First(&k, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *knowledgeRepo) Update(k *models.KnowledgeEntry) error {
	tx := r.db.Save(k)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repositories.ErrStaleData
	}
	return nil
}

func (r *knowledgeRepo) Delete(id string) error {
	return r.db.Delete(&models.KnowledgeEntry{}, "id = ?", id).Error
}

func (r *knowledgeRepo) List() ([]*models.KnowledgeEntry, error) {
	var entries []*models.KnowledgeEntry
	if err := r.db.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *knowledgeRepo) GetByIndustry(industry string) ([]*models.KnowledgeEntry, error) {
	var entries []*models.KnowledgeEntry
	if err := r.db.Where("industry = ?", industry).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
