package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      taskdomain.Status   `json:"status"`
	Recurrence  taskdomain.Recurrence `json:"recurrence"`
}

type upcomingDateDTO struct {
	Date time.Time `json:"date"`
}

type scheduledTaskDTO struct {
	ID     int64             `json:"id"`
	TaskID int64             `json:"task_id"`
	Status taskdomain.Status `json:"status"`
	Date   time.Time         `json:"date"`
}

type taskDTO struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      taskdomain.Status   `json:"status"`
	Recurrence  taskdomain.Recurrence `json:"recurrence"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence:  task.Recurrence,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func newScheduledTaskDTO(task *taskdomain.ScheduledTask) scheduledTaskDTO {
	return scheduledTaskDTO{
		ID:     task.ID,
		TaskID: task.TaskID,
		Status: task.Status,
		Date:   task.Date,
	}
}
