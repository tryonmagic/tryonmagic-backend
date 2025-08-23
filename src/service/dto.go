package service

import "github.com/mbbipul/tryonmagic-backend/src/model"

type GetAIModelsResponse struct {
	Data  []model.AIModel `json:"data"`
	Count int             `json:"count"`
}

type PaginationResult[T any] struct {
	Results []T   `json:"results"`
	Count   int64 `json:"count"`
}
