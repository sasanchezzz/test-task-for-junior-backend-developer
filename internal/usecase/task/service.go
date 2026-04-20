package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:      normalized.Title,
		Description: normalized.Description,
		Status:     normalized.Status,
		Recurrence: normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:         id,
		Title:      normalized.Title,
		Description: normalized.Description,
		Status:     normalized.Status,
		Recurrence: normalized.Recurrence,
		UpdatedAt:  s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetUpcomingDates(ctx context.Context, id int64, count int) ([]taskdomain.DateInfo, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if count <= 0 || count > 100 {
		return nil, fmt.Errorf("%w: count must be between 1 and 100", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task.Recurrence.Type == taskdomain.RecurrenceNone {
		return nil, fmt.Errorf("%w: task has no recurrence", ErrInvalidInput)
	}

	return task.Recurrence.GenerateUpcomingDates(s.now(), count), nil
}

func (s *Service) CreateScheduledDates(ctx context.Context, input CreateScheduledDatesInput) ([]taskdomain.ScheduledTask, error) {
	if input.TaskID <= 0 {
		return nil, fmt.Errorf("%w: invalid task id", ErrInvalidInput)
	}
	if len(input.Dates) == 0 {
		return nil, fmt.Errorf("%w: dates are required", ErrInvalidInput)
	}

	_, err := s.repo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, err
	}

	for _, date := range input.Dates {
		if date.IsZero() {
			return nil, fmt.Errorf("%w: invalid date value", ErrInvalidInput)
		}
	}

	return s.repo.CreateScheduledTasks(ctx, input.TaskID, input.Dates)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if !input.Recurrence.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid recurrence", ErrInvalidInput)
	}

	var startDate time.Time
	if input.Recurrence.StartDate != nil {
		startDate = *input.Recurrence.StartDate
	} else {
		startDate = time.Now().UTC()
	}

	if err := taskdomain.ValidateStartDate(startDate); err != nil {
		return CreateInput{}, err
	}

	if input.Recurrence.EndDate != nil && input.Recurrence.EndDate.Before(startDate) {
		return CreateInput{}, fmt.Errorf("%w: end_date cannot be before start_date", ErrInvalidInput)
	}

	// Warn about invalid months for monthly_on_day
	if input.Recurrence.Type == taskdomain.RecurrenceMonthlyOnDay {
		invalidMonths := input.Recurrence.InvalidMonths()
		if len(invalidMonths) > 0 {
			// We don't reject, just let it through — the GenerateUpcomingDates will skip those months
		}
	}

	input.Recurrence.StartDate = &startDate

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if !input.Recurrence.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid recurrence", ErrInvalidInput)
	}

	var startDate time.Time
	if input.Recurrence.StartDate != nil {
		startDate = *input.Recurrence.StartDate
	} else {
		startDate = time.Now().UTC()
	}

	if err := taskdomain.ValidateStartDate(startDate); err != nil {
		return UpdateInput{}, err
	}

	if input.Recurrence.EndDate != nil && input.Recurrence.EndDate.Before(startDate) {
		return UpdateInput{}, fmt.Errorf("%w: end_date cannot be before start_date", ErrInvalidInput)
	}

	input.Recurrence.StartDate = &startDate

	return input, nil
}
