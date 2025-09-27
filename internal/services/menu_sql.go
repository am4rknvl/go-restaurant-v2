package services

import (
	"context"
	"database/sql"
	"errors"

	"restaurant-system/internal/models"
)

type MenuCategoryDTO struct {
	Name  string            `json:"name"`
	Items []models.MenuItem `json:"items"`
}

type MenuSQLService struct {
	db *sql.DB
}

func NewMenuSQLService(db *sql.DB) *MenuSQLService {
	return &MenuSQLService{db: db}
}

// GetQRMenu groups items by category; restaurantID/tableID are accepted for route shape.
func (s *MenuSQLService) GetQRMenu(ctx context.Context, restaurantID string, tableID string, lang string) ([]MenuCategoryDTO, error) {
	// Fetch distinct categories
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT category FROM menu_items WHERE available = TRUE ORDER BY category ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	// Fetch items per category
	result := make([]MenuCategoryDTO, 0, len(categories))
	for _, cat := range categories {
		itemRows, err := s.db.QueryContext(ctx, "SELECT id, name, description, price, category, available, image_url, special_notes, name_am, description_am FROM menu_items WHERE available = TRUE AND category = $1 ORDER BY name ASC", cat)
		if err != nil {
			return nil, err
		}
		var items []models.MenuItem
		for itemRows.Next() {
			var it models.MenuItem
			var name, desc, nameAm, descAm sql.NullString
			if err := itemRows.Scan(&it.ID, &name, &desc, &it.Price, &it.Category, &it.Available, &it.ImageURL, &it.SpecialNotes, &nameAm, &descAm); err != nil {
				itemRows.Close()
				return nil, err
			}
			// choose language
			if lang == "am" && nameAm.Valid {
				it.Name = nameAm.String
			} else if name.Valid {
				it.Name = name.String
			}
			if lang == "am" && descAm.Valid {
				it.Description = descAm.String
			} else if desc.Valid {
				it.Description = desc.String
			}
			items = append(items, it)
		}
		itemRows.Close()
		result = append(result, MenuCategoryDTO{Name: cat, Items: items})
	}
	return result, nil
}

// CRUD for items
func (s *MenuSQLService) CreateItem(ctx context.Context, it *models.MenuItem) error {
	if it.Name == "" || it.Category == "" || it.Price <= 0 {
		return errors.New("invalid item")
	}
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO menu_items (id, name, description, price, category, available) VALUES ($1,$2,$3,$4,$5,$6)",
		it.ID, it.Name, it.Description, it.Price, it.Category, it.Available,
	)
	return err
}

func (s *MenuSQLService) UpdateItem(ctx context.Context, it *models.MenuItem) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE menu_items SET name=$1, description=$2, price=$3, category=$4, available=$5 WHERE id=$6",
		it.Name, it.Description, it.Price, it.Category, it.Available, it.ID,
	)
	return err
}

func (s *MenuSQLService) DeleteItem(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM menu_items WHERE id=$1", id)
	return err
}

// Fetch average rating for an item
func (s *MenuSQLService) GetAverageRating(ctx context.Context, menuItemID string) (float64, error) {
	var avg sql.NullFloat64
	if err := s.db.QueryRowContext(ctx, "SELECT AVG(rating) FROM reviews WHERE menu_item_id=$1", menuItemID).Scan(&avg); err != nil {
		return 0, err
	}
	if avg.Valid {
		return avg.Float64, nil
	}
	return 0, nil
}

// Favorite operations
func (s *MenuSQLService) AddFavorite(ctx context.Context, fav *models.Favorite) error {
	if fav.ID == "" {
		return errors.New("id required")
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO favorites (id, account_id, menu_item_id, created_at) VALUES ($1,$2,$3,now())", fav.ID, fav.AccountID, fav.MenuItemID)
	return err
}

func (s *MenuSQLService) RemoveFavorite(ctx context.Context, accountID, menuItemID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM favorites WHERE account_id=$1 AND menu_item_id=$2", accountID, menuItemID)
	return err
}

// Reviews
func (s *MenuSQLService) CreateReview(ctx context.Context, r *models.Review) error {
	if r.ID == "" {
		return errors.New("id required")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO reviews (id, account_id, menu_item_id, rating, comment, created_at) VALUES ($1,$2,$3,$4,$5,now())", r.ID, r.AccountID, r.MenuItemID, r.Rating, r.Comment)
	return err
}

// Categories CRUD
func (s *MenuSQLService) CreateCategory(ctx context.Context, cat *models.Category) error {
	if cat.Name == "" {
		return errors.New("name required")
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO categories (id, name, description, created_at, updated_at) VALUES ($1,$2,$3,now(),now())", cat.ID, cat.Name, cat.Description)
	return err
}

func (s *MenuSQLService) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, nil
}

func (s *MenuSQLService) UpdateCategory(ctx context.Context, cat *models.Category) error {
	_, err := s.db.ExecContext(ctx, "UPDATE categories SET name=$1, description=$2, updated_at=now() WHERE id=$3", cat.Name, cat.Description, cat.ID)
	return err
}

func (s *MenuSQLService) DeleteCategory(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM categories WHERE id=$1", id)
	return err
}
