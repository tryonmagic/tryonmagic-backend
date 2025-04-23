package grpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/mbbipul/prodigyai-backend/src/config"
	"github.com/mbbipul/prodigyai-backend/src/model"
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
	task := new(model.Task)
	task.ID = taskInfo.TaskId
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
