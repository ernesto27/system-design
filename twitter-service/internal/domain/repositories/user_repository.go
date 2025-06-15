package repositories

import (
	"context"
	"errors"

	"twitterservice/internal/domain/entities"
	"twitterservice/internal/infrastructure/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository interface defines user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.DB,
	}
}

// CreateUser creates a new user in the database
func (r *userRepository) CreateUser(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserByID retrieves a user by ID
func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *userRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user in the database
func (r *userRepository) UpdateUser(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser deletes a user from the database
func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.User{}, id).Error
}
