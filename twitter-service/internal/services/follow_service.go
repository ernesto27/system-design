package services

import (
	"context"
	"errors"
	"fmt"

	"twitterservice/internal/domain/entities"
	"twitterservice/internal/domain/repositories"

	"github.com/google/uuid"
)

// FollowService handles follow-related business logic
type FollowService struct {
	followRepo repositories.FollowRepository
	userRepo   repositories.UserRepository
}

// NewFollowService creates a new follow service
func NewFollowService(followRepo repositories.FollowRepository, userRepo repositories.UserRepository) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// FollowUserRequest represents a follow user request
type FollowUserRequest struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
}

// FollowUserResponse represents a follow user response
type FollowUserResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	IsPrivate bool   `json:"is_private,omitempty"`
}

// GetFollowersRequest represents a get followers request
type GetFollowersRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// GetFollowersResponse represents a get followers response
type GetFollowersResponse struct {
	Users   []*entities.User `json:"users"`
	Total   int64            `json:"total"`
	HasMore bool             `json:"has_more"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// FollowUser creates a follow relationship between users
func (s *FollowService) FollowUser(ctx context.Context, req *FollowUserRequest) (*FollowUserResponse, error) {
	// Validate input
	if req.FollowerID == uuid.Nil || req.FollowingID == uuid.Nil {
		return nil, errors.New("invalid user IDs")
	}

	// Check if trying to follow themselves
	if req.FollowerID == req.FollowingID {
		return nil, errors.New("users cannot follow themselves")
	}

	// Check if both users exist
	follower, err := s.userRepo.GetUserByID(ctx, req.FollowerID)
	if err != nil {
		return nil, fmt.Errorf("error getting follower: %w", err)
	}
	if follower == nil {
		return nil, errors.New("follower not found")
	}

	following, err := s.userRepo.GetUserByID(ctx, req.FollowingID)
	if err != nil {
		return nil, fmt.Errorf("error getting user to follow: %w", err)
	}
	if following == nil {
		return nil, errors.New("user to follow not found")
	}

	// Check if user to follow is active
	if !following.IsActive {
		return nil, errors.New("cannot follow inactive user")
	}

	// Check if already following
	isFollowing, err := s.followRepo.IsFollowing(ctx, req.FollowerID, req.FollowingID)
	if err != nil {
		return nil, fmt.Errorf("error checking follow status: %w", err)
	}
	if isFollowing {
		return &FollowUserResponse{
			Success: false,
			Message: "Already following this user",
		}, nil
	}

	// Create follow relationship
	err = s.followRepo.FollowUser(ctx, req.FollowerID, req.FollowingID)
	if err != nil {
		return nil, fmt.Errorf("error creating follow relationship: %w", err)
	}

	response := &FollowUserResponse{
		Success: true,
		Message: "Successfully followed user",
	}

	// If the account is private, note that in the response
	if following.IsPrivate {
		response.IsPrivate = true
		response.Message = "Follow request sent to private account"
	}

	return response, nil
}

// UnfollowUser removes a follow relationship between users
func (s *FollowService) UnfollowUser(ctx context.Context, req *FollowUserRequest) (*FollowUserResponse, error) {
	// Validate input
	if req.FollowerID == uuid.Nil || req.FollowingID == uuid.Nil {
		return nil, errors.New("invalid user IDs")
	}

	// Check if follow relationship exists
	isFollowing, err := s.followRepo.IsFollowing(ctx, req.FollowerID, req.FollowingID)
	if err != nil {
		return nil, fmt.Errorf("error checking follow status: %w", err)
	}
	if !isFollowing {
		return &FollowUserResponse{
			Success: false,
			Message: "Not following this user",
		}, nil
	}

	// Remove follow relationship
	err = s.followRepo.UnfollowUser(ctx, req.FollowerID, req.FollowingID)
	if err != nil {
		return nil, fmt.Errorf("error removing follow relationship: %w", err)
	}

	return &FollowUserResponse{
		Success: true,
		Message: "Successfully unfollowed user",
	}, nil
}

// GetFollowers returns the followers of a user
func (s *FollowService) GetFollowers(ctx context.Context, req *GetFollowersRequest) (*GetFollowersResponse, error) {
	// Validate input
	if req.UserID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	// Set default limit and offset
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get followers
	followers, err := s.followRepo.GetFollowers(ctx, req.UserID, req.Limit+1, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("error getting followers: %w", err)
	}

	// Get total count
	totalCount, err := s.followRepo.GetFollowerCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting follower count: %w", err)
	}

	// Check if there are more results
	hasMore := len(followers) > req.Limit
	if hasMore {
		followers = followers[:req.Limit]
	}

	return &GetFollowersResponse{
		Users:   followers,
		Total:   totalCount,
		HasMore: hasMore,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// GetFollowing returns the users that a user follows
func (s *FollowService) GetFollowing(ctx context.Context, req *GetFollowersRequest) (*GetFollowersResponse, error) {
	// Validate input
	if req.UserID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	// Set default limit and offset
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get following
	following, err := s.followRepo.GetFollowing(ctx, req.UserID, req.Limit+1, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("error getting following: %w", err)
	}

	// Get total count
	totalCount, err := s.followRepo.GetFollowingCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting following count: %w", err)
	}

	// Check if there are more results
	hasMore := len(following) > req.Limit
	if hasMore {
		following = following[:req.Limit]
	}

	return &GetFollowersResponse{
		Users:   following,
		Total:   totalCount,
		HasMore: hasMore,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// IsFollowing checks if a user is following another user
func (s *FollowService) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	if followerID == uuid.Nil || followingID == uuid.Nil {
		return false, errors.New("invalid user IDs")
	}

	return s.followRepo.IsFollowing(ctx, followerID, followingID)
}

// GetMutualFollows returns users that both users follow
func (s *FollowService) GetMutualFollows(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*entities.User, error) {
	if userID1 == uuid.Nil || userID2 == uuid.Nil {
		return nil, errors.New("invalid user IDs")
	}

	// Set default limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.followRepo.GetMutualFollows(ctx, userID1, userID2, limit, offset)
}
