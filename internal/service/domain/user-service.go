package domain

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"

	"github.com/google/uuid"
)

type IUserService interface {
	FindAll() []model.User
	FindById(Id uuid.UUID) (*model.User, error)
	FindByGoogleEmail(email string) *model.User
	FindByAppleEmail(email string) *model.User
	Save(model.User) model.User
	Update(model.User) error
	Delete(id uuid.UUID) error
}

type UserService struct {
	userRepository repository.IUserRepository
}

func NewUserService(repo repository.IUserRepository) *UserService {
	return &UserService{
		userRepository: repo,
	}
}

func (u *UserService) FindAll() []model.User {
	return u.userRepository.FindAll()
}

func (u *UserService) FindById(Id uuid.UUID) (*model.User, error) {
	user, err := u.userRepository.FindById(Id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) FindByGoogleEmail(email string) *model.User {
	return u.userRepository.FindByGoogleEmail(email)
}
func (u *UserService) FindByAppleEmail(email string) *model.User {
	return u.userRepository.FindByAppleEmail(email)
}

func (u *UserService) Save(user model.User) model.User {
	u.userRepository.Save(user)
	return user
}

func (u *UserService) Update(user model.User) error {
	return u.userRepository.Update(user)
}

func (u *UserService) Delete(id uuid.UUID) error {
	return u.userRepository.Delete(id)
}
