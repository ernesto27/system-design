package controllers

import (
	"fmt"
	"net/http"
	"server/models"
	"server/response"
)

type ProjectStatusController struct {
	ProjectStatusService models.ProjectStatusService
}

func (controller *ProjectStatusController) GetAllProjectStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := controller.ProjectStatusService.GetAll()
	if err != nil {
		fmt.Println(err)
		response.NewWithoutData().WithMessage("Failed to retrieve project statuses").InternalServerError(w)
		return
	}

	response.New(statuses).WithMessage("Project statuses retrieved successfully").Success(w)
}
