package listServices

import (
	"context"

	"dalista/dalista-be/app/repository"
)

// CategoryInterestRow is one row of the category-interest export: for a given
// user and category, how many of their items and lists fall under it, plus how
// many public lists they've starred in that category (Phase 7).
type CategoryInterestRow struct {
	UserID       string `gorm:"column:user_id"`
	Email        string `gorm:"column:email"`
	Phone        string `gorm:"column:phone"`
	Category     string `gorm:"column:category"`
	ItemCount    int    `gorm:"column:item_count"`
	ListCount    int    `gorm:"column:list_count"`
	StarredCount int    `gorm:"column:starred_count"`
}

// CategoryInterest returns, per (user, category), the counts of owned items and
// lists plus starred public lists in that category. This is the full definition
// of "category interest": categories a user engages with either by owning
// items/lists or by starring others' public lists. Admin-scoped (see the
// handler's permission gate).
func CategoryInterest(ctx context.Context) ([]CategoryInterestRow, error) {
	const q = `
		WITH signals AS (
			SELECT owner_id AS user_id, category_id, 'item' AS kind
			FROM items WHERE category_id IS NOT NULL
			UNION ALL
			SELECT owner_id, category_id, 'list' AS kind
			FROM lists WHERE category_id IS NOT NULL
			UNION ALL
			SELECT s.user_id, l.category_id, 'starred' AS kind
			FROM public_list_stars s
			JOIN lists l ON l.id = s.list_id
			WHERE l.category_id IS NOT NULL
		)
		SELECT o.user_id,
		       COALESCE(u.email, '') AS email,
		       COALESCE(u.phone, '') AS phone,
		       c.name AS category,
		       SUM(CASE WHEN o.kind = 'item' THEN 1 ELSE 0 END) AS item_count,
		       SUM(CASE WHEN o.kind = 'list' THEN 1 ELSE 0 END) AS list_count,
		       SUM(CASE WHEN o.kind = 'starred' THEN 1 ELSE 0 END) AS starred_count
		FROM signals o
		JOIN users u ON u.id = o.user_id
		JOIN categories c ON c.id = o.category_id
		GROUP BY o.user_id, u.email, u.phone, c.name
		ORDER BY o.user_id, c.name`

	var rows []CategoryInterestRow
	err := repository.GetDB().WithContext(ctx).Raw(q).Scan(&rows).Error
	return rows, err
}
