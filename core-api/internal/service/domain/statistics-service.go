package domain

import (
	"SmartSpend/internal/repository"
	"time"
)

type IStatisticsService interface {
	FindTotalIncomeAndExpense(userId string, from time.Time, to time.Time) (float32, float32, error)
	FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error)
	FindTotalSpentPerMonth(userId string, from time.Time, to time.Time) (map[int32]float32, error)
	FindAverage(userId string, from time.Time, to time.Time) (float32, float32, error)
}

type StatisticsService struct {
	statisticsRepository repository.IStatisticsRepository
}

func NewStatisticsService(statisticsRepository repository.IStatisticsRepository) *StatisticsService {
	return &StatisticsService{statisticsRepository: statisticsRepository}
}

func (s *StatisticsService) FindTotalIncomeAndExpense(userId string, from time.Time, to time.Time) (float32, float32, error) {
	return s.statisticsRepository.FindTotalIncomeAndExpense(userId, from, to)
}

func (s *StatisticsService) FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error) {
	return s.statisticsRepository.FindPercentageSpentPerCategory(userId, from, to)
}

func (s *StatisticsService) FindTotalSpentPerMonth(userId string, from time.Time, to time.Time) (map[int32]float32, error) {
	return s.statisticsRepository.FindTotalSpentPerMonth(userId, from, to)
}

func (s *StatisticsService) FindAverage(userId string, from time.Time, to time.Time) (float32, float32, error) {
	return s.statisticsRepository.FindAverage(userId, from, to)
}
