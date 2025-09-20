package domain

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	"time"
)

type ISavingService interface {
	FindAll(userId string, from time.Time, to time.Time) []model.Saving
	FindById(id int64, userId string) (*model.Saving, error)
	Save(saving *model.Saving) error
	Delete(id int64, userId string) error
}

type SavingService struct {
	savingRepository repository.ISavingRepository
}

func NewSavingService(repo repository.ISavingRepository) *SavingService {
	return &SavingService{
		savingRepository: repo,
	}
}

func (t *SavingService) FindAll(userId string, from time.Time, to time.Time) []model.Saving {
	return t.savingRepository.FindAll(userId, from, to)
}
func (t *SavingService) FindById(id int64, userId string) (*model.Saving, error) {
	return t.savingRepository.FindById(id, userId)
}
func (t *SavingService) Save(saving *model.Saving) error {
	return t.savingRepository.Save(*saving)
}
func (t *SavingService) Delete(id int64, userId string) error {
	return t.savingRepository.Delete(id, userId)
}
