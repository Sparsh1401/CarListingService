package repository

import (
	"database/sql"

	"github.com/Sparsh1401/car_listing_service/models"
	"github.com/lib/pq"
)

type CarRepository interface {
	GetAll() ([]models.Car, error)
	GetByID(id int) (*models.Car, error)
	Create(car *models.Car) error
	Update(car *models.Car) error
	Delete(id int) error
	FindExistingLinks(links []string) (map[string]bool, error)
	InsertBatch(cars []models.Car) (int, error)
}

type carRepository struct {
	db *sql.DB
}

func NewCarRepository(db *sql.DB) CarRepository {
	return &carRepository{db: db}
}

func (r *carRepository) GetAll() ([]models.Car, error) {
	query := "SELECT id, title, price, currency, year, location, mileage, link, created_at, updated_at FROM cars ORDER BY created_at DESC"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []models.Car
	for rows.Next() {
		var car models.Car
		if err := rows.Scan(&car.ID, &car.Title, &car.Price, &car.Currency, &car.Year, &car.Location, &car.Mileage, &car.Link, &car.CreatedAt, &car.UpdatedAt); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	return cars, rows.Err()
}

func (r *carRepository) GetByID(id int) (*models.Car, error) {
	query := "SELECT id, title, price, currency, year, location, mileage, link, created_at, updated_at FROM cars WHERE id = $1"
	var car models.Car
	err := r.db.QueryRow(query, id).Scan(
		&car.ID, &car.Title, &car.Price, &car.Currency, &car.Year,
		&car.Location, &car.Mileage, &car.Link, &car.CreatedAt, &car.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &car, nil
}

func (r *carRepository) Create(car *models.Car) error {
	query := `
		INSERT INTO cars (title, price, currency, year, location, mileage, link)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		car.Title, car.Price, car.Currency, car.Year, car.Location, car.Mileage, car.Link,
	).Scan(&car.ID, &car.CreatedAt, &car.UpdatedAt)
}

func (r *carRepository) Update(car *models.Car) error {
	query := `
		UPDATE cars
		SET title = $1, price = $2, currency = $3, year = $4,
		    location = $5, mileage = $6, link = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`
	return r.db.QueryRow(
		query,
		car.Title, car.Price, car.Currency, car.Year,
		car.Location, car.Mileage, car.Link, car.ID,
	).Scan(&car.UpdatedAt)
}

func (r *carRepository) Delete(id int) error {
	query := "DELETE FROM cars WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *carRepository) FindExistingLinks(links []string) (map[string]bool, error) {
	if len(links) == 0 {
		return make(map[string]bool), nil
	}

	query := "SELECT link FROM cars WHERE link = ANY($1)"
	rows, err := r.db.Query(query, pq.Array(links))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingLinks := make(map[string]bool)
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		existingLinks[link] = true
	}

	return existingLinks, rows.Err()
}

func (r *carRepository) InsertBatch(cars []models.Car) (int, error) {
	if len(cars) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO cars (title, price, location, mileage, link)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (link) DO NOTHING
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	insertedCount := 0
	for _, car := range cars {
		result, err := stmt.Exec(car.Title, car.Price, car.Location, car.Mileage, car.Link)
		if err != nil {
			continue
		}

		rows, _ := result.RowsAffected()
		insertedCount += int(rows)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return insertedCount, nil
}
