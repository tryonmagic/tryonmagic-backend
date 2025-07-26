package service

import "github.com/mbbipul/tryonmagic-backend/src/model"

type GetAIModelsResponse struct {
	Data  []model.AIModel `json:"data"`
	Count int             `json:"count"`
}
