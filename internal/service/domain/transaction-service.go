package domain

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"

	"github.com/google/uuid"
)

type ITransactionService interface {
	FindAll(userId uuid.UUID) []model.Transaction
	FindById(Id int64, userId uuid.UUID) (*model.Transaction, error)
	Save(transaction *model.Transaction) error
	Delete(transactionId int64, userId uuid.UUID) error
}

type TransactionService struct {
	transactionRepository repository.ITransactionRepository
}

func NewTransactionService(repo repository.ITransactionRepository) *TransactionService {
	return &TransactionService{
		transactionRepository: repo,
	}
}

func (t *TransactionService) FindAll(userId uuid.UUID) []model.Transaction {
	return t.transactionRepository.FindAll(userId)
}
func (t *TransactionService) FindById(id int64, userId uuid.UUID) (*model.Transaction, error) {
	return t.transactionRepository.FindById(id, userId)
}
func (t *TransactionService) Save(transaction *model.Transaction) error {
	return t.transactionRepository.Save(*transaction)
}
func (t *TransactionService) Delete(transactionId int64, userId uuid.UUID) error {
	return t.transactionRepository.Delete(transactionId, userId)
}
