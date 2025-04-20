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
		response.NewWithoutData().WithMessage("Invalid JSON data: " + err.Error()).BadRequest(w)
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
			response.NewWithoutData().WithMessage(commentErr.Message).InternalServerError(w)
			return
		}
		fmt.Println("Error creating comment:", err)
		response.NewWithoutData().WithMessage("Failed to create comment: " + err.Error()).InternalServerError(w)
		return
	}

	response.New(createdComment).WithMessage("Comment created successfully").Success(w)
}

func (ctrl *Comment) GetProjectComments(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.NewWithoutData().WithMessage("Invalid project ID").BadRequest(w)
		return
	}

	comments, err := ctrl.CommentService.GetProjectComments(projectID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			response.NewWithoutData().WithMessage(commentErr.Message).InternalServerError(w)
			return
		}
		fmt.Println("Error fetching comments:", err)
		response.NewWithoutData().WithMessage("Failed to fetch comments: " + err.Error()).InternalServerError(w)
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
		response.NewWithoutData().WithMessage("Invalid comment ID").BadRequest(w)
		return
	}

	err = ctrl.CommentService.Delete(commentID, userID)
	if err != nil {
		if commentErr, ok := err.(*models.CommentError); ok {
			response.NewWithoutData().WithMessage(commentErr.Message).InternalServerError(w)
			return
		}
		fmt.Println("Error deleting comment:", err)
		response.NewWithoutData().WithMessage("Failed to delete comment: " + err.Error()).InternalServerError(w)
		return
	}

	response.NewWithoutData().WithMessage("Comment deleted successfully").Success(w)
}
