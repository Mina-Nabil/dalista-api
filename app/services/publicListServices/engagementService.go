package publicListServices

import (
	"context"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
)

// EngagementRow is one public list in the engagement table, with its owner's
// contact fields and star count, for the dashboard.
type EngagementRow struct {
	ListID    string `gorm:"column:list_id" json:"list_id"`
	Name      string `gorm:"column:name" json:"name"`
	OwnerID   string `gorm:"column:owner_id" json:"owner_id"`
	Email     string `gorm:"column:email" json:"email"`
	Phone     string `gorm:"column:phone" json:"phone"`
	StarCount int    `gorm:"column:star_count" json:"star_count"`
}

// Engagement aggregates public-list engagement for the dashboard: totals plus
// the top lists by star count.
type Engagement struct {
	TotalPublicLists int             `json:"total_public_lists"`
	TotalStars       int             `json:"total_stars"`
	TopLists         []EngagementRow `json:"top_lists"`
}

// EngagementStats returns the public-list engagement summary. topN caps the
// "top lists by stars" table.
func EngagementStats(ctx context.Context, topN int) (Engagement, error) {
	if topN <= 0 {
		topN = 10
	}
	db := repository.GetDB().WithContext(ctx)
	out := Engagement{TopLists: []EngagementRow{}}

	type totals struct {
		Lists int `gorm:"column:lists"`
		Stars int `gorm:"column:stars"`
	}
	var t totals
	if err := db.Model(&listModels.List{}).
		Select("COUNT(*) AS lists, COALESCE(SUM(star_count), 0) AS stars").
		Where("is_public = TRUE").Scan(&t).Error; err != nil {
		return out, err
	}
	out.TotalPublicLists = t.Lists
	out.TotalStars = t.Stars

	err := db.Table("lists l").
		Select(`l.id AS list_id, l.name AS name, l.owner_id AS owner_id,
		         COALESCE(u.email,'') AS email, COALESCE(u.phone,'') AS phone,
		         l.star_count AS star_count`).
		Joins("JOIN users u ON u.id = l.owner_id").
		Where("l.is_public = TRUE").
		Order("l.star_count DESC, l.created_at DESC").
		Limit(topN).Scan(&out.TopLists).Error
	if err != nil {
		return out, err
	}
	return out, nil
}
