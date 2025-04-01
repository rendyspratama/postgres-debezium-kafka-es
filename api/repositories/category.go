package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/rendyspratama/digital-discovery/api/config"
	"github.com/rendyspratama/digital-discovery/api/models"
)

type CategoryRepository interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(category *models.Category) error
	DeleteCategory(id int) error
	GetCategoriesWithPagination(page, perPage int) ([]models.Category, int, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository() CategoryRepository {
	return &categoryRepository{
		db: config.GetDB(),
	}
}

func (r *categoryRepository) GetAllCategories() ([]models.Category, error) {
	rows, err := r.db.Query(`
		SELECT id, name, status, created_at, updated_at 
		FROM categories 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *categoryRepository) GetCategoryByID(id int) (*models.Category, error) {
	var c models.Category
	err := r.db.QueryRow(`
		SELECT id, name, status, created_at, updated_at 
		FROM categories 
		WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Status, &c.CreatedAt, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) CreateCategory(category *models.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRow(`
		INSERT INTO categories (name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, category.Name, category.Status, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)

	if err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) UpdateCategory(category *models.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	category.UpdatedAt = time.Now()

	result, err := r.db.Exec(`
		UPDATE categories 
		SET name = $1, status = $2, updated_at = $3
		WHERE id = $4
	`, category.Name, category.Status, category.UpdatedAt, category.ID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("category not found")
	}

	return nil
}

func (r *categoryRepository) DeleteCategory(id int) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("category not found")
	}

	return nil
}

func (r *categoryRepository) GetCategoriesWithPagination(page, perPage int) ([]models.Category, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	rows, err := r.db.Query(`
		SELECT id, name, status, created_at, updated_at 
		FROM categories 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		categories = append(categories, c)
	}
	return categories, total, nil
}
