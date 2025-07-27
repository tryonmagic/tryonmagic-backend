package response

import "github.com/mbbipul/tryonmagic-backend/src/model"

type Common struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SuccessWithUser struct {
	Code    int        `json:"code"`
	Status  string     `json:"status"`
	Message string     `json:"message"`
	User    model.User `json:"user"`
}

type GenericSuccessResponse struct {
	Code    int    `json:"code" example:"201"`
	Message string `json:"message" example:"Register successfully"`
	Data    any    `json:"data"`
}

type GenericErrorResponse struct {
	Code    int    `json:"code" example:"201"`
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Register successfully"`
	Errors  []any  `json:"errors"`
}

type SuccessWithTokens struct {
	Code    int        `json:"code"`
	Status  string     `json:"status"`
	Message string     `json:"message"`
	User    model.User `json:"user"`
	Tokens  Tokens     `json:"tokens"`
}

type PaginateData[T any] struct {
	Results      []T   `json:"results"`
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	TotalPages   int64 `json:"totalPages"`
	TotalResults int64 `json:"totalResults"`
}

type SuccessWithPaginate[T any] struct {
	Code    int             `json:"code"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    PaginateData[T] `json:"data"`
}

type ErrorDetails struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors"`
}
