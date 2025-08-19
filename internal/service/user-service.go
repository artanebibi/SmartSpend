package service

import (
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	"github.com/google/uuid"
)

type IUserService interface {
	FindAll() []model.User
	FindById(Id uuid.UUID) *model.User
	//FindByEmail(email string) *model.User
	Save(model.User) model.User
	//Update(model.User) model.User
	//Delete(model.User)
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

func (u *UserService) FindById(Id uuid.UUID) *model.User {
	user, err := u.userRepository.FindById(Id)
	if err != nil {
		return nil
	}
	return user
}

//func (u *UserService) FindByEmail(email string) *model.User {
//	return u.userRepository.FindByEmail(email)
//}

func (u *UserService) Save(user model.User) model.User {
	u.userRepository.Save(user)
	return user
}

//func (u *UserService) Update(user model.User) model.User {
//	u.userRepository.Update(user)
//	return user
//}
//
//func (u *UserService) Delete(user model.User) {
//	u.userRepository.Delete(user)
//}
