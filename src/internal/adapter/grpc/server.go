package grpc

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/mbbipul/tryonmagic-backend/src/config"
	socket "github.com/mbbipul/tryonmagic-backend/src/internal/adapter/socketio"
	"github.com/mbbipul/tryonmagic-backend/src/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"google.golang.org/grpc"
)

type server struct {
	TaskServiceServer
	DB *gorm.DB
}

func (s *server) UpdateTaskStatus(_ context.Context, in *UpdateTaskStatusRequest) (*UpdateTaskStatusRequest, error) {
	log.Printf("Received task update request: %v", in)
	task, err := s.UpdateTask(in)
	if err != nil {
		log.Printf("Failed to update task: %v", err)
		return nil, err
	}
	log.Printf("Task updated successfully: %v", task)
	return &UpdateTaskStatusRequest{}, nil
}

func StartServer(db *gorm.DB) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	RegisterTaskServiceServer(s, &server{
		DB: db,
	})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (t *server) UpdateTask(taskInfo *UpdateTaskStatusRequest) (*model.Task, error) {
	var resultData *datatypes.JSON
	if taskInfo.ResultData != nil {
		if taskInfo.ResultData != nil {
			jsonBytes, err := taskInfo.ResultData.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal ResultData: %w", err)
			}
			resultData = (*datatypes.JSON)(&jsonBytes)
		}
	}

	fmt.Printf("Updating task: %s, status: %s, resultData: %v", taskInfo.TaskId, taskInfo.Status, resultData)
	taskId, err := uuid.Parse(taskInfo.TaskId)
	if err != nil {
		fmt.Printf("Failed to parse task ID: %s, error: %v", taskInfo.TaskId, err)
		return nil, err
	}
	task := new(model.Task)
	task.ID = taskId
	task.Status = model.TaskStatus(taskInfo.Status)
	task.ResultData = resultData
	task.Error = &taskInfo.Error
	task.RetryCount = int(taskInfo.RetryCount)
	task.StartedAt = parseTime(taskInfo.StartedAt)
	task.CompletedAt = parseTime(taskInfo.CompletedAt)

	if err := t.DB.WithContext(context.Background()).Model(task).Where("id = ?", taskInfo.TaskId).Updates(model.Task{
		Status:      task.Status,
		ResultData:  task.ResultData,
		Error:       task.Error,
		RetryCount:  task.RetryCount,
		StartedAt:   task.StartedAt,
		CompletedAt: task.CompletedAt,
	}).Error; err != nil {
		fmt.Printf("Failed to update task: %+v", err)
		return nil, err
	}

	if taskInfo.Status == string(model.TaskStatusSucceeded) && taskInfo.Type == string(model.TaskTypeGenerateAIModel) {
		var resultMap map[string]interface{}
		if resultData != nil {
			if err := json.Unmarshal([]byte(*resultData), &resultMap); err != nil {
				fmt.Printf("Failed to unmarshal resultData: %+v", err)
				return nil, err
			}
		}
		outputImageUrl, ok := resultMap["outputImageUrl"].(string)
		if !ok {
			fmt.Printf("outputImageUrl key not found or invalid in resultData")
			return nil, fmt.Errorf("outputImageUrl key not found or invalid in resultData")
		}
		resolution := datatypes.JSON{}
		if width, ok := resultMap["width"].(float64); ok {
			if height, ok := resultMap["height"].(float64); ok {
				resolution = datatypes.JSON(fmt.Sprintf(`{"width": %d, "height": %d}`, int(width), int(height)))
			} else {
				fmt.Printf("height key not found or invalid in resultData")
			}
		} else {
			fmt.Printf("width key not found or invalid in resultData")
		}

		if err := t.DB.WithContext(context.Background()).Model(&model.AIModel{}).Where("task_id = ?", taskInfo.TaskId).Updates(model.AIModel{
			ModelImageUrl: &outputImageUrl,
			Resolution:    resolution,
		}).Error; err != nil {
			fmt.Printf("Failed to update AI model: %+v", err)
			return nil, err
		}
	}

	userId := taskInfo.UserId
	err = socket.EmitEventToUser(userId, "task-progress", map[string]any{
		"id":         taskInfo.TaskId,
		"type":       task.Type,
		"status":     task.Status,
		"resultData": task.ResultData,
	})
	if err != nil {
		fmt.Printf("Failed to emit event to user: %+v", err)
	}
	fmt.Printf("Emitted event to user: %s, task ID: %s,type: %s status: %s", userId, task.ID, taskInfo.Type, task.Status)
	return task, nil
}

func parseTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}
	t, err := time.Parse(time.UTC.String(), timeStr)
	if err != nil {
		return nil
	}
	return &t
}
