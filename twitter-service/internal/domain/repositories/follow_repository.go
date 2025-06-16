package repositories

import (
	"context"
	"errors"

	"twitterservice/internal/domain/entities"
	"twitterservice/internal/infrastructure/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FollowRepository interface defines follow data operations
type FollowRepository interface {
	// Follow operations
	FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error
	UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)

	// Get relationships
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)

	// Get counts
	GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// Mutual follows
	GetMutualFollows(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*entities.User, error)
}

// followRepository implements FollowRepository interface
type followRepository struct {
	db *gorm.DB
}

// NewFollowRepository creates a new follow repository
func NewFollowRepository() FollowRepository {
	return &followRepository{
		db: database.DB,
	}
}

// FollowUser creates a follow relationship
func (r *followRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	// Check if already following
	var existingFollow entities.Follow
	err := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existingFollow).Error
	if err == nil {
		return errors.New("already following this user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new follow relationship
	follow := &entities.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	return r.db.WithContext(ctx).Create(follow).Error
}

// UnfollowUser removes a follow relationship
func (r *followRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&entities.Follow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("follow relationship not found")
	}
	return nil
}

// IsFollowing checks if a user is following another user
func (r *followRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFollowers returns users who follow the specified user
func (r *followRepository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.id, users.google_id, users.email, users.username, users.display_name, 
			users.avatar_url, users.bio, users.location, users.website, 
			users.follower_count, users.following_count, users.post_count,
			users.is_verified, users.is_private, users.is_active,
			users.created_at, users.updated_at`).
		Joins("INNER JOIN follows ON follows.follower_id = users.id").
		Where("follows.following_id = ? AND users.is_active = ?", userID, true).
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, err
}

// GetFollowing returns users that the specified user follows
func (r *followRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.id, users.google_id, users.email, users.username, users.display_name, 
			users.avatar_url, users.bio, users.location, users.website, 
			users.follower_count, users.following_count, users.post_count,
			users.is_verified, users.is_private, users.is_active,
			users.created_at, users.updated_at`).
		Joins("INNER JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ? AND users.is_active = ?", userID, true).
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, err
}

// GetFollowerCount returns the number of followers for a user
func (r *followRepository) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Follow{}).Where("following_id = ?", userID).Count(&count).Error
	return count, err
}

// GetFollowingCount returns the number of users that a user follows
func (r *followRepository) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

// GetMutualFollows returns users that both users follow (mutual connections)
func (r *followRepository) GetMutualFollows(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.id, users.google_id, users.email, users.username, users.display_name, 
			users.avatar_url, users.bio, users.location, users.website, 
			users.follower_count, users.following_count, users.post_count,
			users.is_verified, users.is_private, users.is_active,
			users.created_at, users.updated_at`).
		Joins("INNER JOIN follows f1 ON f1.following_id = users.id").
		Joins("INNER JOIN follows f2 ON f2.following_id = users.id").
		Where("f1.follower_id = ? AND f2.follower_id = ? AND users.is_active = ?", userID1, userID2, true).
		Order("users.username").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, err
}
