package application

import (
	"SmartSpend/internal/domain/dto"
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	"fmt"
	"time"
)

type IApplicationSavingService interface {
	FindAll(userId string, from time.Time, to time.Time) []dto.SavingDto
	FindById(id int64, userId string) (*dto.SavingDto, error)
	Save(savingDto *dto.SavingDto) error
	CreateOrUpdate(savingDto *dto.SavingDto, userId string) (error, string)
	Delete(savingDto *dto.SavingDto, userId string) error
}

type ApplicationSavingService struct {
	savingRepository repository.ISavingRepository
}

func NewApplicationSavingService(repo repository.ISavingRepository) *ApplicationSavingService {
	return &ApplicationSavingService{
		savingRepository: repo,
	}
}

func mapToDtoSaving(s model.Saving) dto.SavingDto {
	return dto.SavingDto{
		ID:     s.ID,
		Amount: &s.Amount,
		From:   &s.From,
		To:     &s.To,
	}
}

func mapToModelSaving(dto *dto.SavingDto) model.Saving {
	return model.Saving{
		ID:     dto.ID,
		Amount: *dto.Amount,
		From:   *dto.From,
		To:     *dto.To,
	}
}

func (s *ApplicationSavingService) FindAll(userId string, from time.Time, to time.Time) []dto.SavingDto {
	savings := s.savingRepository.FindAll(userId, from, to)
	result := make([]dto.SavingDto, len(savings))
	for i, tx := range savings {
		result[i] = mapToDtoSaving(tx)
	}
	return result
}

func (s *ApplicationSavingService) FindById(id int64, userId string) (*dto.SavingDto, error) {
	tx, err := s.savingRepository.FindById(id, userId)
	if err != nil {
		return nil, err
	}
	savingDto := mapToDtoSaving(*tx)
	return &savingDto, nil
}

func (s *ApplicationSavingService) Save(savingDto *dto.SavingDto) error {
	tx := mapToModelSaving(savingDto)
	// Fill in date if missing
	if tx.From.IsZero() {
		tx.From = time.Now()
	}
	return s.savingRepository.Save(tx)
}

func (s *ApplicationSavingService) CreateOrUpdate(savingDto *dto.SavingDto, userId string) (error, string) {
	if savingDto.ID != 0 { // update
		existing, err := s.savingRepository.FindById(savingDto.ID, userId)
		if err != nil {
			return err, "Saving not found"
		}

		saving := *existing

		if savingDto.Amount != nil {
			saving.Amount = *savingDto.Amount
		}
		if savingDto.From != nil {
			saving.From = *savingDto.From
		}
		if savingDto.To != nil {
			saving.To = *savingDto.To
		}

		err = s.savingRepository.Update(saving, saving.ID)
		if err != nil {
			return err, err.Error()
		}
		return nil, fmt.Sprintf("Saving with id %d updated successfully", saving.ID)
	} else {
		if savingDto.Amount == nil {
			return fmt.Errorf("amount is required"), "Amount is required for new saving"
		}
		if savingDto.From == nil {
			return fmt.Errorf("from is required"), "From date is required for new saving"
		}
		if savingDto.To == nil {
			return fmt.Errorf("to is required"), "To date is required for new saving"
		}

		saving := model.Saving{
			OwnerId: userId,
			Amount:  *savingDto.Amount,
			From:    *savingDto.From,
			To:      *savingDto.To,
		}

		err := s.savingRepository.Save(saving)
		if err != nil {
			return err, err.Error()
		}
		return nil, "Saving successfully created."
	}
}

func (s *ApplicationSavingService) Delete(savingDto *dto.SavingDto, userId string) error {
	return s.savingRepository.Delete(savingDto.ID, userId)
}
