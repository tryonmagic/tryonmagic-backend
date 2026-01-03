package validation

import "github.com/mbbipul/tryonmagic-backend/src/model"

type CreateUser struct {
	FirstName string     `json:"firstName" validate:"required,max=50" example:"fake name"`
	LastName  string     `json:"lastName" validate:"required,max=50" example:"fake name"`
	Email     string     `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password  string     `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
	Role      model.Role `json:"role" validate:"required,oneof=user admin,max=50" example:"user"`
}

type UpdateUser struct {
	FirstName string `json:"firstName,omitempty" validate:"omitempty,max=50" example:"fake name"`
	LastName  string `json:"lastName,omitempty" validate:"omitempty,max=50" example:"fake name"`
	Email     string `json:"email" validate:"omitempty,email,max=50" example:"fake@example.com"`
	Password  string `json:"password,omitempty" validate:"omitempty,min=8,max=20,password" example:"password1"`
}

type UpdatePassOrVerify struct {
	Password      string `json:"password,omitempty" validate:"omitempty,min=8,max=20,password" example:"password1"`
	EmailVerified bool   `json:"emailVerified" swaggerignore:"true" validate:"omitempty,boolean"`
}

type QueryUser struct {
	Page   int    `validate:"omitempty,number,max=50"`
	Limit  int    `validate:"omitempty,number,max=50"`
	Search string `validate:"omitempty,max=50"`
}

type GetTasksQueryParams struct {
	BasePaginationQueryParams
	DateFrom   *string          `json:"dateFrom" example:"2025-01-01"`
	DateTo     *string          `json:"dateTo" example:"2025-01-31"`
	TaskStatus model.TaskStatus `json:"taskStatus" validate:"required,oneof=PENDING QUEUED PROCESSING SUCCEEDED FAILED" example:"PENDING"`
	TaskType   model.TaskType   `json:"taskType" validate:"required,oneof=REMOVE_BACKGROUND REPLACE_BACKGROUND GENERATE_AI_MODEL VTON UPSCALE ENHANCE EXPAND_IMAGE EDIT_IMAGE PRODUCT_IMAGE_EDITING" example:"REMOVE_BACKGROUND"`
}
