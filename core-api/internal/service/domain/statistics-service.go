package domain

import (
	"SmartSpend/internal/repository"
	"time"
)

type IStatisticsService interface {
	FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error)
}

type StatisticsService struct {
	statisticsRepository repository.IStatisticsRepository
}

func NewStatisticsService(statisticsRepository repository.IStatisticsRepository) *StatisticsService {
	return &StatisticsService{statisticsRepository: statisticsRepository}
}

func (s *StatisticsService) FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error) {
	return s.statisticsRepository.FindPercentageSpentPerCategory(userId, from, to)
}
