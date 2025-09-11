package application

import (
	"SmartSpend/internal/domain/dto"
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	"fmt"
	"time"
)

type IApplicationTransactionService interface {
	FindAll(userId string, from time.Time, to time.Time) []dto.TransactionDto
	FindById(id int64, userId string) (*dto.TransactionDto, error)
	Save(transactionDto *dto.TransactionDto) error
	CreateOrUpdate(transactionDto *dto.TransactionDto, userId string) (error, string)
	Delete(transactionDto *dto.TransactionDto, userId string) error
}

type ApplicationTransactionService struct {
	transactionRepository repository.ITransactionRepository
}

func NewApplicationTransactionService(repo repository.ITransactionRepository) *ApplicationTransactionService {
	return &ApplicationTransactionService{
		transactionRepository: repo,
	}
}

func mapToDto(t model.Transaction) dto.TransactionDto {
	return dto.TransactionDto{
		ID:         t.ID,
		Title:      &t.Title,
		Price:      &t.Price,
		DateMade:   &t.DateMade,
		CategoryId: t.CategoryId,
		Type:       &t.Type,
	}
}

func mapToModel(dto *dto.TransactionDto) model.Transaction {
	return model.Transaction{
		ID:         dto.ID,
		Title:      *dto.Title,
		Price:      *dto.Price,
		DateMade:   *dto.DateMade,
		CategoryId: dto.CategoryId,
		Type:       *dto.Type,
	}
}

func (s *ApplicationTransactionService) FindAll(userId string, from time.Time, to time.Time) []dto.TransactionDto {
	transactions := s.transactionRepository.FindAll(userId, from, to)
	result := make([]dto.TransactionDto, len(transactions))
	for i, tx := range transactions {
		result[i] = mapToDto(tx)
	}
	return result
}

func (s *ApplicationTransactionService) FindById(id int64, userId string) (*dto.TransactionDto, error) {
	tx, err := s.transactionRepository.FindById(id, userId)
	if err != nil {
		return nil, err
	}
	transactionDto := mapToDto(*tx)
	return &transactionDto, nil
}

func (s *ApplicationTransactionService) Save(transactionDto *dto.TransactionDto) error {
	tx := mapToModel(transactionDto)
	// Fill in date if missing
	if tx.DateMade.IsZero() {
		tx.DateMade = time.Now()
	}
	return s.transactionRepository.Save(tx)
}

func (s *ApplicationTransactionService) CreateOrUpdate(transactionDto *dto.TransactionDto, userId string) (error, string) {
	if transactionDto.ID != 0 { // update
		existing, err := s.transactionRepository.FindById(transactionDto.ID, userId)
		if err != nil {
			return err, "Transaction not found"
		}

		transaction := *existing

		// Only update fields that are provided
		if transactionDto.Title != nil {
			transaction.Title = *transactionDto.Title
		}
		if transactionDto.Price != nil {
			transaction.Price = *transactionDto.Price
		}
		if transactionDto.DateMade != nil {
			transaction.DateMade = *transactionDto.DateMade
		}
		if transactionDto.CategoryId != nil {
			transaction.CategoryId = transactionDto.CategoryId
		}
		if transactionDto.Type != nil {
			transaction.Type = *transactionDto.Type
		}

		err = s.transactionRepository.Update(transaction, transaction.ID)
		if err != nil {
			return err, err.Error()
		}
		return nil, fmt.Sprintf("Transaction with id %d updated successfully", transaction.ID)
	} else {
		if transactionDto.Title == nil || *transactionDto.Title == "" {
			return fmt.Errorf("title is required"), "Title is required for new transaction"
		}
		if transactionDto.Price == nil {
			return fmt.Errorf("price is required"), "Price is required for new transaction"
		}
		if transactionDto.Type == nil {
			return fmt.Errorf("type is required"), "Type is required for new transaction"
		}

		transaction := model.Transaction{
			OwnerId:  userId,
			Title:    *transactionDto.Title,
			Price:    *transactionDto.Price,
			Type:     *transactionDto.Type,
			DateMade: time.Now(),
		}

		if transactionDto.DateMade != nil {
			transaction.DateMade = *transactionDto.DateMade
		}
		if transactionDto.CategoryId != nil {
			transaction.CategoryId = transactionDto.CategoryId
		}

		err := s.transactionRepository.Save(transaction)
		if err != nil {
			return err, err.Error()
		}
		return nil, "Transaction successfully created."
	}
}

func (s *ApplicationTransactionService) Delete(transactionDto *dto.TransactionDto, userId string) error {
	return s.transactionRepository.Delete(transactionDto.ID, userId)
}
