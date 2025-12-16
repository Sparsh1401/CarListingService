package services

import (
	"log"

	"github.com/Sparsh1401/car_listing_service/models"
	"github.com/Sparsh1401/car_listing_service/repository"
)

type CarService interface {
	GetAllCars() ([]models.Car, error)
	GetCarByID(id int) (*models.Car, error)
	CreateCar(car *models.Car) error
	UpdateCar(car *models.Car) error
	DeleteCar(id int) error
	ScrapeAndStoreCars() (int, error)
}

type carService struct {
	repo repository.CarRepository
}

func NewCarService(repo repository.CarRepository) CarService {
	return &carService{repo: repo}
}

func (s *carService) GetAllCars() ([]models.Car, error) {
	return s.repo.GetAll()
}

func (s *carService) GetCarByID(id int) (*models.Car, error) {
	return s.repo.GetByID(id)
}

func (s *carService) CreateCar(car *models.Car) error {
	return s.repo.Create(car)
}

func (s *carService) UpdateCar(car *models.Car) error {
	return s.repo.Update(car)
}

func (s *carService) DeleteCar(id int) error {
	return s.repo.Delete(id)
}

func (s *carService) ScrapeAndStoreCars() (int, error) {
	resultsChan := make(chan []models.Car)
	doneChan := make(chan error)

	go func() {
		err := ScrapeCars(resultsChan)
		close(resultsChan)
		doneChan <- err
	}()

	totalCount := 0

	for batch := range resultsChan {
		if len(batch) == 0 {
			continue
		}

		links := make([]string, len(batch))
		for i, car := range batch {
			links[i] = car.Link
		}

		existingLinks, err := s.repo.FindExistingLinks(links)
		if err != nil {
			log.Printf("Error checking existing links: %v", err)
			continue
		}

		var newCars []models.Car
		for _, car := range batch {
			if !existingLinks[car.Link] {
				newCars = append(newCars, car)
			}
		}

		if len(newCars) == 0 {
			continue
		}

		count, err := s.repo.InsertBatch(newCars)
		if err != nil {
			log.Printf("Error inserting batch: %v", err)
			continue
		}

		totalCount += count
	}

	err := <-doneChan
	return totalCount, err
}
