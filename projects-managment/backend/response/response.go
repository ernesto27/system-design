package response

import (
	"encoding/json"
	"net/http"
)

const SuccessMessage = "Successful request"
const InternalErrorMessage = "Internal server error"
const BadRequestMessage = "Bad request"
const UnauthorizedMessage = "Unauthorized"

type Response struct {
	Status        int    `json:"status"`
	Message       string `json:"message"`
	Data          any    `json:"data"`
	customMessage string
}

func New(data any) *Response {
	responseAPI := &Response{
		Status:  http.StatusOK,
		Message: SuccessMessage,
		Data:    data,
	}

	return responseAPI
}

func NewWithoutData() *Response {
	return New(nil)
}

func (r *Response) Success(w http.ResponseWriter) {
	if r.customMessage != "" {
		r.Message = r.customMessage
	}
	r.WriteJSON(w, http.StatusOK, nil)
}

func (r *Response) InternalServerError(w http.ResponseWriter) {
	r.Status = http.StatusInternalServerError
	r.Message = InternalErrorMessage
	r.WriteJSON(w, http.StatusInternalServerError, nil)
}

func (r *Response) BadRequest(w http.ResponseWriter) {
	r.Status = http.StatusBadRequest
	r.Message = BadRequestMessage
	if r.customMessage != "" {
		r.Message = r.customMessage
	}
	r.WriteJSON(w, http.StatusBadRequest, nil)
}

func (r *Response) WithMessage(message string) *Response {
	r.customMessage = message
	return r
}

func (r *Response) Unauthorized(w http.ResponseWriter) {
	r.Status = http.StatusUnauthorized
	r.Message = UnauthorizedMessage
	r.WriteJSON(w, http.StatusUnauthorized, nil)
}

func (r *Response) WriteJSON(w http.ResponseWriter, status int, headers http.Header) error {
	js, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
