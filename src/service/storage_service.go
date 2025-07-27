package service

import (
	"context"

	"github.com/mbbipul/tryonmagic-backend/src/utils"
	"github.com/mbbipul/tryonmagic-backend/src/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StorageService interface {
	SaveFile(c *fiber.Ctx, req *validation.SaveFile) (*model.File, error)
	GetFilesByUserID(c context.Context, userID uuid.UUID, queryParams *validation.GetFilesQueryParams) (*GetFilesResponse, error)
}

type storageService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewStorageService(db *gorm.DB, validate *validator.Validate) StorageService {
	return &storageService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (s *storageService) SaveFile(c *fiber.Ctx, req *validation.SaveFile) (*model.File, error) {
	s.Log.Infof("Saving file: %+v", req)
	if err := s.Validate.Struct(req); err != nil {
		s.Log.Errorf("Validation error: %+v", err)
		return nil, err
	}
	user := c.Locals("user").(*model.User)
	file := &model.File{
		FullPath: req.FullPath,
		Key:      req.Key,
		URL:      req.URL,
		Name:     req.Name,
		Type:     req.Type,
		Size:     req.Size,
		Width:    req.Width,
		Height:   req.Height,
		Ext:      req.Ext,
		Mime:     req.Mime,
		UserID:   user.ID,
	}
	if err := s.DB.WithContext(c.Context()).Create(file).Error; err != nil {
		s.Log.Errorf("Failed to save file: %+v", err)
		return nil, err
	}
	return file, nil
}

func (s *storageService) GetFilesByUserID(c context.Context, userID uuid.UUID, queryParams *validation.GetFilesQueryParams) (*GetFilesResponse, error) {

	var files []model.File
	query := s.DB.WithContext(c).Model(&model.File{}).Where("user_id = ?", userID)
	if queryParams.Search != nil {
		query = query.Where("name LIKE ? OR ext LIKE ? OR type LIKE ?", "%"+*queryParams.Search+"%", "%"+*queryParams.Search+"%", "%"+*queryParams.Search+"%")
	}
	if queryParams.Types != nil {
		query = query.Where("ext IN ?", *queryParams.Types)
	}
	query = constructGetFilesByUserIDSort(queryParams, query)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		s.Log.Errorf("Failed to get files count: %+v", err)
		return nil, err
	}
	offset := GetPaginationOffset(queryParams.Page, queryParams.PageSize)
	query = query.Offset(offset).Limit(queryParams.PageSize)
	if err := query.Find(&files).Error; err != nil {
		s.Log.Errorf("Failed to get files: %+v", err)
		return nil, err
	}
	return &GetFilesResponse{
		Results: files,
		Count:   count,
	}, nil
}

func constructGetFilesByUserIDSort(queryParams *validation.GetFilesQueryParams, query *gorm.DB) *gorm.DB {
	if queryParams.SortBy != nil {
		if queryParams.SortBy != nil && *queryParams.SortBy != "" {
			allowedSortFields := map[string]string{
				"id":        "id",
				"name":      "name",
				"ext":       "ext",
				"type":      "type",
				"size":      "size",
				"width":     "width",
				"height":    "height",
				"createdAt": "created_at",
			}

			if field, ok := allowedSortFields[*queryParams.SortBy]; ok {
				if queryParams.SortOrder == validation.SortOrderDesc {
					query = query.Order(field + " DESC")
				} else {
					query = query.Order(field + " ASC")
				}
			}
		} else {
			query = query.Order("files.created_at DESC")
		}
	} else {
		query = query.Order("files.created_at DESC")
	}
	return query
}
