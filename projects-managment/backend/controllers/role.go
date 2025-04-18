package controllers

import (
	"net/http"

	"server/models"
	"server/response"
)

type RoleController struct {
	RoleService models.RoleService
}

func (rc *RoleController) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := rc.RoleService.GetAll()
	if err != nil {
		response.NewWithoutData().WithMessage("Failed to retrieve roles").InternalServerError(w)
		return
	}
	response.New(roles).WithMessage("Roles retrieved successfully").Success(w)
}
