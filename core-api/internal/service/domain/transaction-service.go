package domain

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
)

type ITransactionService interface {
	FindAll(userId string) []model.Transaction
	FindById(Id int64, userId string) (*model.Transaction, error)
	Save(transaction *model.Transaction) error
	Delete(transactionId int64, userId string) error
}

type TransactionService struct {
	transactionRepository repository.ITransactionRepository
}

func NewTransactionService(repo repository.ITransactionRepository) *TransactionService {
	return &TransactionService{
		transactionRepository: repo,
	}
}

func (t *TransactionService) FindAll(userId string) []model.Transaction {
	return t.transactionRepository.FindAll(userId)
}
func (t *TransactionService) FindById(id int64, userId string) (*model.Transaction, error) {
	return t.transactionRepository.FindById(id, userId)
}
func (t *TransactionService) Save(transaction *model.Transaction) error {
	return t.transactionRepository.Save(*transaction)
}
func (t *TransactionService) Delete(transactionId int64, userId string) error {
	return t.transactionRepository.Delete(transactionId, userId)
}
