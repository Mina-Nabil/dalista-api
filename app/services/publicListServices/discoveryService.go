package publicListServices

import (
	"context"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	publicListModels "dalista/dalista-be/app/repository/models/publicListModels"

	"github.com/google/uuid"
)

// Discovery pagination bounds.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// DiscoveryItem is one public list in the feed, with whether the caller has
// starred it.
type DiscoveryItem struct {
	List        listModels.List
	StarredByMe bool
}

// ClampLimit normalizes a requested page size to [1, MaxLimit], defaulting to
// DefaultLimit for non-positive input.
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// DiscoverFeed returns a page of public lists sorted by star count (then
// newest), optionally filtered to a category, plus the total count for
// pagination and a per-row "starred by me" flag.
func DiscoverFeed(ctx context.Context, userID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]DiscoveryItem, int64, error) {
	limit = ClampLimit(limit)
	if offset < 0 {
		offset = 0
	}
	db := repository.GetDB().WithContext(ctx)

	base := db.Model(&listModels.List{}).Where("is_public = TRUE")
	if categoryID != nil {
		base = base.Where("category_id = ?", *categoryID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var lists []listModels.List
	if err := base.Order("star_count DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&lists).Error; err != nil {
		return nil, 0, err
	}

	starred, err := starredAmong(ctx, userID, lists)
	if err != nil {
		return nil, 0, err
	}

	items := make([]DiscoveryItem, 0, len(lists))
	for i := range lists {
		items = append(items, DiscoveryItem{List: lists[i], StarredByMe: starred[lists[i].ID]})
	}
	return items, total, nil
}

// starredAmong returns the set of list IDs (from the given page) that the user
// has starred, in one query.
func starredAmong(ctx context.Context, userID uuid.UUID, lists []listModels.List) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	if len(lists) == 0 {
		return out, nil
	}
	ids := make([]uuid.UUID, 0, len(lists))
	for i := range lists {
		ids = append(ids, lists[i].ID)
	}
	var starredIDs []uuid.UUID
	err := repository.GetDB().WithContext(ctx).
		Model(&publicListModels.PublicListStar{}).
		Where("user_id = ? AND list_id IN ?", userID, ids).
		Pluck("list_id", &starredIDs).Error
	if err != nil {
		return nil, err
	}
	for _, id := range starredIDs {
		out[id] = true
	}
	return out, nil
}
