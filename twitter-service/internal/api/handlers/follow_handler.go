package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"twitterservice/internal/domain/entities"
	"twitterservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FollowHandler handles follow-related HTTP requests
type FollowHandler struct {
	followService *services.FollowService
}

// NewFollowHandler creates a new follow handler
func NewFollowHandler(followService *services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// FollowRequest represents the request body for follow/unfollow
type FollowRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// FollowUser handles following a user
// POST /api/users/:id/follow
func (h *FollowHandler) FollowUser(c *gin.Context) {
	// Get user ID from URL params
	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	fmt.Println(followingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get follower ID from request context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	fmt.Println(userInterface)

	// Extract user ID from the user interface
	user, ok := userInterface.(*entities.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user data in context"})
		return
	}

	followerID := user.ID
	fmt.Println("User ID:", followerID)

	req := &services.FollowUserRequest{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// Call service
	_, err = h.followService.FollowUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{"message": "User followed successfully"})
}

// UnfollowUser handles unfollowing a user
// DELETE /api/users/:id/follow
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	// Get user ID from URL params
	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get follower ID from request context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract user ID from the user interface
	user, ok := userInterface.(*entities.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user data in context"})
		return
	}

	followerID := user.ID

	// Create unfollow request
	req := &services.FollowUserRequest{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// Call service
	_, err = h.followService.UnfollowUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{"message": "User unfollowed successfully"})
}

// GetFollowers handles getting a user's followers
// GET /api/users/:id/followers
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	// Get user ID from URL params
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Create request
	req := &services.GetFollowersRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Call service
	response, err := h.followService.GetFollowers(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetFollowing handles getting users that a user follows
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	// Get user ID from URL params
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Create request
	req := &services.GetFollowersRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Call service
	response, err := h.followService.GetFollowing(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// CheckFollowStatus checks if the current user is following another user
// GET /api/users/:id/follow/status
func (h *FollowHandler) CheckFollowStatus(c *gin.Context) {
	// Get user ID from URL params
	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get follower ID from request context (set by auth middleware)
	followerIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	followerID, ok := followerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in context"})
		return
	}

	// Check follow status
	isFollowing, err := h.followService.IsFollowing(c.Request.Context(), followerID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}

// GetMutualFollows handles getting mutual follows between two users
// GET /api/users/:id/mutual/:otherUserId
func (h *FollowHandler) GetMutualFollows(c *gin.Context) {
	// Get user IDs from URL params
	userID1Str := c.Param("id")
	userID2Str := c.Param("otherUserId")

	userID1, err := uuid.Parse(userID1Str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userID2, err := uuid.Parse(userID2Str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Call service
	mutualFollows, err := h.followService.GetMutualFollows(c.Request.Context(), userID1, userID2, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"mutual_follows": mutualFollows,
		"count":          len(mutualFollows),
		"limit":          limit,
		"offset":         offset,
	})
}
