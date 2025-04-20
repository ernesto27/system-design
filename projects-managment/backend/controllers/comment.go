package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"server/internal"
	"server/models"
	"server/response"

	"github.com/go-chi/chi"
)

type Comment struct {
	CommentService models.CommentService
}

func (ctrl *Comment) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := internal.GetUserIDFromContext(r.Context())
	if !ok {
		response.NewWithoutData().WithMessage("Unauthorized").Unauthorized(w)
		return
	}

	var comment models.Comment

	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		fmt.Println("Error parsing comment JSON:", err)
		response.NewWithoutData().WithMessage("Invalid request format").BadRequest(w)
		return
	}

	if comment.ProjectID == 0 || comment.Content == "" {
		response.NewWithoutData().WithMessage("Project ID and content are required").BadRequest(w)
		return
	}

	comment.UserID = userID
	createdComment, err := ctrl.CommentService.Create(comment)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			fmt.Println("Comment error:", commentErr.Message)
			response.NewWithoutData().WithMessage("Failed to create comment").InternalServerError(w)
		}
		fmt.Println("Error creating comment:", err)
		response.NewWithoutData().WithMessage("Failed to create comment").InternalServerError(w)
		return
	}

	response.New(createdComment).WithMessage("Comment created successfully").Success(w)
}

func (ctrl *Comment) GetProjectComments(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		fmt.Println("Invalid project ID:", err)
		response.NewWithoutData().WithMessage("Invalid project ID").BadRequest(w)
		return
	}

	userID, _ := internal.GetUserIDFromContext(r.Context())

	comments, err := ctrl.CommentService.GetProjectComments(projectID, userID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			fmt.Println("Comment error:", commentErr.Message)
			response.NewWithoutData().WithMessage("Failed to fetch comments").InternalServerError(w)
			return
		}
		fmt.Println("Error fetching comments:", err)
		response.NewWithoutData().WithMessage("Failed to fetch comments").InternalServerError(w)
		return
	}

	response.New(comments).WithMessage("Comments retrieved successfully").Success(w)
}

func (ctrl *Comment) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := internal.GetUserIDFromContext(r.Context())
	if !ok {
		response.NewWithoutData().WithMessage("Unauthorized").Unauthorized(w)
		return
	}

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		fmt.Println("Invalid comment ID:", err)
		response.NewWithoutData().WithMessage("Invalid comment ID").BadRequest(w)
		return
	}

	err = ctrl.CommentService.Delete(commentID, userID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			fmt.Println("Comment error when deleting:", commentErr.Message)
			response.NewWithoutData().WithMessage("Failed to delete comment").InternalServerError(w)
			return
		}
		fmt.Println("Error deleting comment:", err)
		response.NewWithoutData().WithMessage("Failed to delete comment").InternalServerError(w)
		return
	}

	response.NewWithoutData().WithMessage("Comment deleted successfully").Success(w)
}

func (ctrl *Comment) LikeComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := internal.GetUserIDFromContext(r.Context())
	if !ok {
		response.NewWithoutData().WithMessage("Unauthorized").Unauthorized(w)
		return
	}

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		fmt.Println("Invalid comment ID:", err)
		response.NewWithoutData().WithMessage("Invalid comment ID").BadRequest(w)
		return
	}

	err = ctrl.CommentService.LikeComment(commentID, userID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			fmt.Println("Comment error when liking:", commentErr.Message)
			response.NewWithoutData().WithMessage("Failed to like comment").InternalServerError(w)
			return
		}
		fmt.Println("Error liking comment:", err)
		response.NewWithoutData().WithMessage("Failed to like comment").InternalServerError(w)
		return
	}

	response.NewWithoutData().WithMessage("Comment liked successfully").Success(w)
}

func (ctrl *Comment) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := internal.GetUserIDFromContext(r.Context())
	if !ok {
		response.NewWithoutData().WithMessage("Unauthorized").Unauthorized(w)
		return
	}

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		fmt.Println("Invalid comment ID:", err)
		response.NewWithoutData().WithMessage("Invalid comment ID").BadRequest(w)
		return
	}

	err = ctrl.CommentService.UnlikeComment(commentID, userID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			fmt.Println("Comment error when unliking:", commentErr.Message)
			response.NewWithoutData().WithMessage("Failed to unlike comment").InternalServerError(w)
			return
		}
		fmt.Println("Error unliking comment:", err)
		response.NewWithoutData().WithMessage("Failed to unlike comment").InternalServerError(w)
		return
	}

	response.NewWithoutData().WithMessage("Comment unliked successfully").Success(w)
}
