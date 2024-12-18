package api

var (
	IncorrectParameter      = "INCORRECT_PARAMETER"
	ServerError             = "SERVER_ERROR"
	NotFound                = "NOT_FOUND"
	InvalidEmailOrPassword  = "INVALID_EMAIL_OR_PASSWORD"
	UnprocessableEntity     = "UNPROCESSABLE_ENTITY"
	EmailAlreadyExists      = "EMAIL_ALREADY_EXISTS"
	InvalidVerificationCode = "INVALID_VERIFICATION_CODE"
	Forbidden               = "FORBIDDEN"
	Unauthorized            = "UNAUTHORIZED"
	TokenInvalidOrExpired   = "TOKEN_INVALID_OR_EXPIRED"
)

type SuccessResponse[T any] struct {
	Status  string `json:"status"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Error struct {
	Code      string `json:"code"`
	Parameter string `json:"parameter"`
	Message   string `json:"message"`
}

type ErrorResponse struct {
	Errors []*Error `json:"error"`
	Status string   `json:"status"`
}

func NewErrorResponse(errors []*Error) *ErrorResponse {
	return &ErrorResponse{
		Errors: errors,
		Status: "fail",
	}
}

func NewSuccessResponse[T any](data T, message string) *SuccessResponse[T] {
	return &SuccessResponse[T]{
		Data:    data,
		Status:  "success",
		Message: message,
	}
}
