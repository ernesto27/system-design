package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal"
	"server/models"
	"server/response"
)

type User struct {
	UserService models.UserService
	JWTService  internal.JWTService
}
type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const MessageUserOrPasswordEmpty = "email or password is empty"

func (user *User) Login(w http.ResponseWriter, r *http.Request) {
	jsonUser := Request{}
	err := json.NewDecoder(r.Body).Decode(&jsonUser)
	if err != nil {
		fmt.Println(err)
		response.NewWithoutData().BadRequest(w)
		return
	}

	if jsonUser.Email == "" || jsonUser.Password == "" {
		fmt.Println(MessageUserOrPasswordEmpty)
		response.NewWithoutData().WithMessage(MessageUserOrPasswordEmpty).BadRequest(w)
		return
	}

	userResp, err := user.UserService.Login(jsonUser.Email, jsonUser.Password)
	if err != nil {
		fmt.Println(err)
		if err == models.ErrUserNotFound {
			fmt.Println("user not found")
			response.NewWithoutData().WithMessage(MessageUserOrPasswordEmpty).BadRequest(w)
			return
		}

		response.NewWithoutData().InternalServerError(w)
		return
	}

	tokens, err := user.JWTService.GenerateTokenPairClient(userResp.ID)
	if err != nil {
		response.NewWithoutData().InternalServerError(w)
		return
	}

	response.New(internal.TokenClienteResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}).Success(w)

}
