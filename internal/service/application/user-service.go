package application

import (
	"SmartSpend/internal/domain/dto"
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/service/domain"

	"github.com/google/uuid"
)

type IUserAppService interface {
	FindAll() []dto.UserDto
	FindById(id uuid.UUID) (*dto.UserDto, error)
	FindByGoogleEmail(email string) *dto.UserDto
	FindByAppleEmail(email string) *dto.UserDto
	Save(user dto.UserDto) dto.UserDto
	Update(userID uuid.UUID, user dto.UpdateUserDto) error
}

type UserAppService struct {
	domainService domain.IUserService
}

func NewUserAppService(ds domain.IUserService) *UserAppService {
	return &UserAppService{
		domainService: ds,
	}
}

func mapToUserDTO(user *model.User) *dto.UserDto {
	if user == nil {
		return nil
	}
	return &dto.UserDto{
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Username:          user.Username,
		GoogleEmail:       user.GoogleEmail,
		AppleEmail:        user.AppleEmail,
		AvatarURL:         user.AvatarURL,
		CreatedAt:         user.CreatedAt,
		Balance:           user.Balance,
		MonthlySavingGoal: user.MonthlySavingGoal,
		PreferredCurrency: user.PreferredCurrency,
	}
}

func mapToUpdateUserDTO(user *model.User) *dto.UpdateUserDto {
	if user == nil {
		return nil
	}
	return &dto.UpdateUserDto{
		FirstName:         &user.FirstName,
		LastName:          &user.LastName,
		Username:          &user.Username,
		AvatarURL:         &user.AvatarURL,
		Balance:           &user.Balance,
		MonthlySavingGoal: &user.MonthlySavingGoal,
		PreferredCurrency: &user.PreferredCurrency,
	}
}

func mapSliceToDTO(users []model.User) []dto.UserDto {
	result := make([]dto.UserDto, len(users))
	for i, u := range users {
		result[i] = *mapToUserDTO(&u)
	}
	return result
}

// Application service methods

func (u *UserAppService) FindAll() []dto.UserDto {
	domainUsers := u.domainService.FindAll()
	return mapSliceToDTO(domainUsers)
}

func (u *UserAppService) FindById(id uuid.UUID) (*dto.UserDto, error) {
	user, err := u.domainService.FindById(id)
	if err != nil {
		return nil, err
	}
	return mapToUserDTO(user), nil
}

func (u *UserAppService) FindByGoogleEmail(email string) *dto.UserDto {
	user := u.domainService.FindByGoogleEmail(email)
	return mapToUserDTO(user)
}

func (u *UserAppService) FindByAppleEmail(email string) *dto.UserDto {
	user := u.domainService.FindByAppleEmail(email)
	return mapToUserDTO(user)
}

func (u *UserAppService) Save(user dto.UserDto) dto.UserDto {
	domainUser := model.User{
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Username:          user.Username,
		GoogleEmail:       user.GoogleEmail,
		AppleEmail:        user.AppleEmail,
		AvatarURL:         user.AvatarURL,
		CreatedAt:         user.CreatedAt,
		Balance:           user.Balance,
		MonthlySavingGoal: user.MonthlySavingGoal,
		PreferredCurrency: user.PreferredCurrency,
	}
	saved := u.domainService.Save(domainUser)
	return *mapToUserDTO(&saved)
}

func (u *UserAppService) Update(id uuid.UUID, newUserUpdate dto.UpdateUserDto) error {
	existing, err := u.domainService.FindById(id)
	if err != nil {
		return err
	}

	if newUserUpdate.FirstName != nil {
		existing.FirstName = *newUserUpdate.FirstName
	}
	if newUserUpdate.LastName != nil {
		existing.LastName = *newUserUpdate.LastName
	}
	if newUserUpdate.Username != nil {
		existing.Username = *newUserUpdate.Username
	}
	if newUserUpdate.AvatarURL != nil {
		existing.AvatarURL = *newUserUpdate.AvatarURL
	}
	if newUserUpdate.Balance != nil {
		existing.Balance = *newUserUpdate.Balance
	}
	if newUserUpdate.MonthlySavingGoal != nil {
		existing.MonthlySavingGoal = *newUserUpdate.MonthlySavingGoal
	}
	if newUserUpdate.PreferredCurrency != nil {
		existing.PreferredCurrency = *newUserUpdate.PreferredCurrency
	}

	return u.domainService.Update(*existing)
}
