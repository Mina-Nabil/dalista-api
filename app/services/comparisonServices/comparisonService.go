package comparisonServices

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"dalista/dalista-be/app/repository"
	comparisonModels "dalista/dalista-be/app/repository/models/comparisonModels"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	listServices "dalista/dalista-be/app/services/listServices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// View is a resolved comparison: the session plus the items the viewer may see,
// in the owner's selection order, and a count of items withheld because the
// viewer lacks access (or the item was deleted).
type View struct {
	Session  *comparisonModels.ComparisonSession
	Items    []listModels.Item
	Withheld int
}

// Create persists a comparison from a manually selected, ordered set of item
// IDs. Every item must be accessible to the creator (a list they can read, or a
// public catalog item); otherwise ErrValidation. Duplicates are removed while
// preserving first-seen order. Fewer than two accessible items is ErrValidation.
func Create(ctx context.Context, ownerID uuid.UUID, itemIDs []uuid.UUID) (*View, error) {
	ordered := dedupe(itemIDs)
	if len(ordered) < 2 {
		return nil, ErrValidation
	}

	items, byID, err := loadItems(ctx, ordered)
	if err != nil {
		return nil, err
	}
	lists, err := loadLists(ctx, items)
	if err != nil {
		return nil, err
	}
	// Creation is strict: every selected item must be accessible to the creator.
	for _, id := range ordered {
		item, ok := byID[id]
		if !ok {
			return nil, ErrValidation
		}
		okAccess, err := canAccessItem(ctx, ownerID, item, lists)
		if err != nil {
			return nil, err
		}
		if !okAccess {
			return nil, ErrValidation
		}
	}

	raw, err := json.Marshal(uuidStrings(ordered))
	if err != nil {
		return nil, err
	}
	session := &comparisonModels.ComparisonSession{OwnerID: ownerID, ItemIDs: string(raw)}
	if err := repository.GetDB().WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	// Return in the stored order (all accessible to the creator).
	view, err := resolve(ctx, ownerID, session)
	if err != nil {
		return nil, err
	}
	return view, nil
}

// Get returns a comparison the caller owns, with its items gated against the
// caller. Owner-scoped: a non-owner gets ErrNotFound.
func Get(ctx context.Context, userID, id uuid.UUID) (*View, error) {
	session, err := getOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return resolve(ctx, userID, session)
}

// GetByToken returns a shared comparison by its share token, with its items
// gated against the viewer — items the viewer can't access are withheld, not
// errored. Sharing grants no new access.
func GetByToken(ctx context.Context, viewerID uuid.UUID, token string) (*View, error) {
	var session comparisonModels.ComparisonSession
	err := repository.GetDB().WithContext(ctx).
		Where("share_token = ?", token).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return resolve(ctx, viewerID, &session)
}

// Share generates (or returns the existing) share token for a comparison the
// caller owns. Owner-only.
func Share(ctx context.Context, userID, id uuid.UUID) (*comparisonModels.ComparisonSession, error) {
	session, err := getOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if session.ShareToken != nil && *session.ShareToken != "" {
		return session, nil
	}
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	if err := repository.GetDB().WithContext(ctx).Model(session).
		Update("share_token", token).Error; err != nil {
		return nil, err
	}
	session.ShareToken = &token
	return session, nil
}

// --- internals ---

func getOwned(ctx context.Context, userID, id uuid.UUID) (*comparisonModels.ComparisonSession, error) {
	var session comparisonModels.ComparisonSession
	err := repository.GetDB().WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, userID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// resolve loads the session's items and gates each against the viewer, keeping
// the owner's selection order and counting withheld (inaccessible or deleted).
func resolve(ctx context.Context, viewerID uuid.UUID, session *comparisonModels.ComparisonSession) (*View, error) {
	ids, err := parseIDs(session.ItemIDs)
	if err != nil {
		return nil, err
	}
	items, byID, err := loadItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	lists, err := loadLists(ctx, items)
	if err != nil {
		return nil, err
	}

	view := &View{Session: session, Items: make([]listModels.Item, 0, len(ids))}
	for _, id := range ids {
		item, ok := byID[id]
		if !ok {
			view.Withheld++ // deleted since the comparison was saved
			continue
		}
		okAccess, err := canAccessItem(ctx, viewerID, item, lists)
		if err != nil {
			return nil, err
		}
		if !okAccess {
			view.Withheld++
			continue
		}
		view.Items = append(view.Items, *item)
	}
	return view, nil
}

// canAccessItem reports whether the viewer may see an item: it's in a public
// list (public catalog), or in a list the viewer can read (owner/editor/viewer).
func canAccessItem(ctx context.Context, viewerID uuid.UUID, item *listModels.Item, lists map[uuid.UUID]listModels.List) (bool, error) {
	if l, ok := lists[item.ListID]; ok && l.IsPublic {
		return true, nil
	}
	return listServices.CanRead(ctx, viewerID, item.ListID)
}

func loadItems(ctx context.Context, ids []uuid.UUID) ([]listModels.Item, map[uuid.UUID]*listModels.Item, error) {
	if len(ids) == 0 {
		return nil, map[uuid.UUID]*listModels.Item{}, nil
	}
	var items []listModels.Item
	if err := repository.GetDB().WithContext(ctx).Preload("Attributes").Preload("Images").Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	byID := make(map[uuid.UUID]*listModels.Item, len(items))
	for i := range items {
		byID[items[i].ID] = &items[i]
	}
	return items, byID, nil
}

func loadLists(ctx context.Context, items []listModels.Item) (map[uuid.UUID]listModels.List, error) {
	out := map[uuid.UUID]listModels.List{}
	if len(items) == 0 {
		return out, nil
	}
	seen := map[uuid.UUID]bool{}
	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if !seen[items[i].ListID] {
			seen[items[i].ListID] = true
			ids = append(ids, items[i].ListID)
		}
	}
	var lists []listModels.List
	if err := repository.GetDB().WithContext(ctx).Where("id IN ?", ids).Find(&lists).Error; err != nil {
		return nil, err
	}
	for i := range lists {
		out[lists[i].ID] = lists[i]
	}
	return out, nil
}

func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == (uuid.UUID{}) || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func parseIDs(raw string) ([]uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	var strs []string
	if err := json.Unmarshal([]byte(raw), &strs); err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			continue // skip malformed ids rather than failing the whole read
		}
		out = append(out, id)
	}
	return out, nil
}

func newToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
