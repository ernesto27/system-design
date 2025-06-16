package services

import (
	"fmt"
	"time"
	"twitterservice/internal/domain/entities"
	"twitterservice/internal/domain/repositories"

	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(userID uuid.UUID, content string) (*entities.Post, error)
	GetPost(id uuid.UUID) (*entities.Post, error)
	GetUserPosts(userID uuid.UUID, limit int) ([]*entities.Post, error)
	UpdatePost(id uuid.UUID, userID uuid.UUID, content string) (*entities.Post, error)
	DeletePost(id uuid.UUID, userID uuid.UUID) error
}

type postService struct {
	postRepo repositories.PostRepository
}

func NewPostService(postRepo repositories.PostRepository) PostService {
	return &postService{
		postRepo: postRepo,
	}
}

func (s *postService) CreatePost(userID uuid.UUID, content string) (*entities.Post, error) {
	post := &entities.Post{
		ID:        uuid.New(),
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}

	fmt.Printf("Creating post: ID=%s, UserID=%s, Content=%s\n", post.ID, post.UserID, content)

	err := s.postRepo.Create(post)
	if err != nil {
		fmt.Printf("Error in repository Create: %v\n", err)
		return nil, err
	}

	fmt.Printf("Post created successfully: %s\n", post.ID)
	return post, nil
}

func (s *postService) GetPost(id uuid.UUID) (*entities.Post, error) {
	return s.postRepo.GetByID(id)
}

func (s *postService) GetUserPosts(userID uuid.UUID, limit int) ([]*entities.Post, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	return s.postRepo.GetByUserID(userID, limit)
}

func (s *postService) UpdatePost(id uuid.UUID, userID uuid.UUID, content string) (*entities.Post, error) {
	// First check if the post exists and belongs to the user
	existingPost, err := s.postRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existingPost == nil {
		return nil, ErrPostNotFound
	}
	if existingPost.UserID != userID {
		return nil, ErrUnauthorized
	}

	existingPost.Content = content
	existingPost.UpdatedAt = time.Now()

	err = s.postRepo.Update(existingPost)
	if err != nil {
		return nil, err
	}

	return existingPost, nil
}

func (s *postService) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	// First check if the post exists and belongs to the user
	existingPost, err := s.postRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existingPost == nil {
		return ErrPostNotFound
	}
	if existingPost.UserID != userID {
		return ErrUnauthorized
	}

	return s.postRepo.Delete(id)
}

// Service errors
var (
	ErrPostNotFound = &ServiceError{Code: "POST_NOT_FOUND", Message: "Post not found"}
	ErrUnauthorized = &ServiceError{Code: "UNAUTHORIZED", Message: "You are not authorized to perform this action"}
)

type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
