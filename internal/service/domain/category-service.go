package domain

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
)

type ICategoryService interface {
	FindAll() []model.Category
}

type CategoryService struct {
	categoryRepository repository.ICategoryRepository
}

func NewCategoryService(repo repository.ICategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepository: repo,
	}
}

func (c *CategoryService) FindAll() []model.Category {
	return c.categoryRepository.FindAll()
}
