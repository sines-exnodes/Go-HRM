package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// InviteListFilter is the admin listing query. Status is a derived
// pseudo-enum (pending|accepted|expired|revoked) — translated into the
// equivalent WHERE clauses in List().
type InviteListFilter struct {
	Email    string
	Status   string
	Page     int
	PageSize int
}

// InviteRepository owns the invites aggregate. Soft-delete IS used here
// (Revoke flips is_deleted=true) so reads use the NotDeleted scope
// UNLESS the caller is explicitly looking up a token (some flows want
// to find a revoked invite to return a "this invite was revoked" error
// — though Phase 9 keeps it simple: revoked = 404).
type InviteRepository interface {
	Create(ctx context.Context, inv *models.Invite) error
	Update(ctx context.Context, inv *models.Invite) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Invite, error)
	FindByToken(ctx context.Context, token string) (*models.Invite, error)
	FindPendingByEmail(ctx context.Context, email string) (*models.Invite, error)
	List(ctx context.Context, f InviteListFilter) ([]models.Invite, int64, error)
}

type inviteRepository struct{ db *gorm.DB }

// NewInviteRepository constructs the repo.
func NewInviteRepository(db *gorm.DB) InviteRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *inviteRepository) Create(ctx context.Context, inv *models.Invite) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *inviteRepository) Update(ctx context.Context, inv *models.Invite) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *inviteRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Invite{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

func (r *inviteRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Invite, error) {
	var inv models.Invite
	err := r.base(ctx).
		Preload("Inviter").
		Where("id = ?", id).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepository) FindByToken(ctx context.Context, token string) (*models.Invite, error) {
	var inv models.Invite
	err := r.base(ctx).
		Where("token = ?", token).
		First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

// FindPendingByEmail finds the (at-most-one) pending invite for an email.
// Used by Create to short-circuit the partial-unique-constraint violation
// with a clean 409.
func (r *inviteRepository) FindPendingByEmail(ctx context.Context, email string) (*models.Invite, error) {
	var inv models.Invite
	err := r.base(ctx).
		Where("LOWER(email) = LOWER(?) AND accepted_at IS NULL", strings.TrimSpace(email)).
		First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepository) List(ctx context.Context, f InviteListFilter) ([]models.Invite, int64, error) {
	q := r.base(ctx).Model(&models.Invite{})

	if e := strings.TrimSpace(f.Email); e != "" {
		q = q.Where("LOWER(email) LIKE LOWER(?)", "%"+e+"%")
	}

	// Status derivation — pending/accepted/expired/revoked. revoked
	// is_deleted=true, so the NotDeleted scope above already excludes
	// it. To list revoked invites we'd need a separate code path
	// (out of scope for Phase 9 — drafted as a follow-up).
	now := time.Now().UTC()
	switch strings.ToLower(strings.TrimSpace(f.Status)) {
	case "pending":
		q = q.Where("accepted_at IS NULL AND expires_at > ?", now)
	case "accepted":
		q = q.Where("accepted_at IS NOT NULL")
	case "expired":
		q = q.Where("accepted_at IS NULL AND expires_at <= ?", now)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 20
	}

	var rows []models.Invite
	err := q.
		Preload("Inviter").
		Order("created_at DESC").
		Limit(size).Offset((page - 1) * size).
		Find(&rows).Error
	return rows, total, err
}
