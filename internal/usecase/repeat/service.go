package repeat

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/teambition/rrule-go"
)

type Service struct {
	periodRepo PeriodRepository
	repeatRepo RepeatTaskRepository
	taskRepo   TaskRepository
	now        func() time.Time
}

func NewService(
	periodRepo PeriodRepository,
	repeatRepo RepeatTaskRepository,
	taskRepo TaskRepository,
) *Service {
	return &Service{
		periodRepo: periodRepo,
		repeatRepo: repeatRepo,
		taskRepo:   taskRepo,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// =============== Periods ===============

func (s *Service) CreatePeriod(ctx context.Context, input CreatePeriodInput) (*taskdomain.Period, error) {
	if err := validatePeriodInput(input.Code, input.Title); err != nil {
		return nil, err
	}
	now := s.now()
	p := &taskdomain.Period{
		Code:          strings.TrimSpace(input.Code),
		Title:         strings.TrimSpace(input.Title),
		RRULETemplate: input.RRULETemplate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.periodRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetPeriodByID(ctx context.Context, id int64) (*taskdomain.Period, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.periodRepo.GetByID(ctx, id)
}

func (s *Service) UpdatePeriod(ctx context.Context, id int64, input UpdatePeriodInput) (*taskdomain.Period, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if err := validatePeriodInput(input.Code, input.Title); err != nil {
		return nil, err
	}
	p, err := s.periodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Code = strings.TrimSpace(input.Code)
	p.Title = strings.TrimSpace(input.Title)
	p.RRULETemplate = input.RRULETemplate
	p.UpdatedAt = s.now()
	if err := s.periodRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePeriod(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.periodRepo.Delete(ctx, id)
}

func (s *Service) ListPeriods(ctx context.Context) ([]taskdomain.Period, error) {
	return s.periodRepo.List(ctx)
}

// =============== RepeatTasks ===============

func (s *Service) CreateRepeatTask(ctx context.Context, input CreateRepeatTaskInput) (*taskdomain.RepeatTask, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if err := validateRepeatTaskInput(input.PeriodID, input.RRULE, input.CustomDates); err != nil {
		return nil, err
	}

	// Если указан PeriodID — проверяем существование и берём rrule из периода
	var resolvedRRULE *string
	if input.PeriodID != nil {
		period, err := s.periodRepo.GetByID(ctx, *input.PeriodID)
		if err != nil {
			return nil, fmt.Errorf("%w: period not found", ErrInvalidInput)
		}
		resolvedRRULE = period.RRULETemplate
	} else {
		resolvedRRULE = input.RRULE
	}

	// Нормализуем даты: убираем время, оставляем только дату в UTC
	var customDates taskdomain.DateArray
	for _, t := range input.CustomDates {
		customDates = append(customDates, time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
	}

	now := s.now()
	rt := &taskdomain.RepeatTask{
		Title:           strings.TrimSpace(input.Title),
		Description:     strings.TrimSpace(input.Description),
		Status:          input.Status,
		PeriodID:        input.PeriodID,
		RRULE:           resolvedRRULE,
		CustomDates:     customDates,
		Enabled:         input.Enabled,
		LastGeneratedAt: nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repeatRepo.Create(ctx, rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (s *Service) GetRepeatTaskByID(ctx context.Context, id int64) (*taskdomain.RepeatTask, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repeatRepo.GetByID(ctx, id)
}

func (s *Service) UpdateRepeatTask(ctx context.Context, id int64, input UpdateRepeatTaskInput) (*taskdomain.RepeatTask, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	rt, err := s.repeatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Собираем итоговые значения правила повторения для валидации
	// Берём текущие значения как базу, затем применяем изменения из input
	newPeriodID := rt.PeriodID
	newRRULE := rt.RRULE
	var newCustomDates []time.Time
	for _, d := range rt.CustomDates {
		newCustomDates = append(newCustomDates, d)
	}

	// Если хоть одно поле правила передано — сбрасываем все три и берём из input
	ruleChanged := input.PeriodID != nil || input.RRULE != nil || input.CustomDates != nil
	if ruleChanged {
		newPeriodID = input.PeriodID
		newRRULE = input.RRULE
		newCustomDates = input.CustomDates
	}

	if ruleChanged {
		if err := validateRepeatTaskInput(newPeriodID, newRRULE, newCustomDates); err != nil {
			return nil, err
		}
	}

	// Если сменился PeriodID — обновляем rrule из нового периода
	if newPeriodID != nil {
		period, err := s.periodRepo.GetByID(ctx, *newPeriodID)
		if err != nil {
			return nil, fmt.Errorf("%w: period not found", ErrInvalidInput)
		}
		newRRULE = period.RRULETemplate
	}

	// Обновляем поля шаблона задачи
	if input.Title != nil {
		rt.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		rt.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		rt.Status = *input.Status
	}
	if input.Enabled != nil {
		rt.Enabled = *input.Enabled
	}

	// Применяем обновлённое правило
	if ruleChanged {
		rt.PeriodID = newPeriodID
		rt.RRULE = newRRULE

		var customDates taskdomain.DateArray
		for _, t := range newCustomDates {
			customDates = append(customDates, time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
		}
		rt.CustomDates = customDates
	}

	rt.UpdatedAt = s.now()

	if err := s.repeatRepo.Update(ctx, rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (s *Service) DeleteRepeatTask(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repeatRepo.Delete(ctx, id)
}

func (s *Service) ListActiveRepeatTasks(ctx context.Context) ([]taskdomain.RepeatTask, error) {
	return s.repeatRepo.ListActive(ctx)
}

func (s *Service) ListRepeatTasks(ctx context.Context) ([]taskdomain.RepeatTask, error) {
	return s.repeatRepo.List(ctx)
}

func (s *Service) GetTasksWithRepeatInfo(ctx context.Context, limit, offset int) ([]taskdomain.TaskWithRepeatInfo, error) {
	return s.repeatRepo.GetTasksWithRepeatInfo(ctx, limit, offset)
}

// =============== Создание задачи + шаблона за раз ===============

func (s *Service) CreateTaskWithRepeat(ctx context.Context, taskInput CreateTaskInput, repeatInput CreateRepeatTaskInput) (*taskdomain.Task, *taskdomain.RepeatTask, error) {
	// Создаём шаблон повторения
	rt, err := s.CreateRepeatTask(ctx, repeatInput)
	if err != nil {
		return nil, nil, fmt.Errorf("create repeat task: %w", err)
	}

	// Создаём первую задачу от шаблона
	now := s.now()
	task := &taskdomain.Task{
		Title:        strings.TrimSpace(taskInput.Title),
		Description:  strings.TrimSpace(taskInput.Description),
		Status:       taskInput.Status,
		RepeatTaskID: &rt.ID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if task.Title == "" {
		task.Title = rt.Title
	}
	if task.Status == "" {
		task.Status = rt.Status
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, nil, fmt.Errorf("create task: %w", err)
	}

	return task, rt, nil
}

// =============== Генерация задач ===============

func (s *Service) GenerateTasksForRepeatTask(ctx context.Context, repeatTaskID int64, from, to time.Time) (int, error) {
	if repeatTaskID <= 0 {
		return 0, fmt.Errorf("%w: invalid repeat task id", ErrInvalidInput)
	}

	rt, err := s.repeatRepo.GetByID(ctx, repeatTaskID)
	if err != nil {
		return 0, err
	}
	if !rt.Enabled {
		return 0, fmt.Errorf("%w: repeat task is disabled", ErrInvalidInput)
	}

	fromUTC := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toUTC := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, time.UTC)

	var occurrences []time.Time

	if len(rt.CustomDates) > 0 {
		// Режим конкретных дат — просто фильтруем по диапазону
		for _, date := range rt.CustomDates {
			date = date.UTC()
			if !date.Before(fromUTC) && !date.After(toUTC) {
				occurrences = append(occurrences, date)
			}
		}
	} else {
		// Режим RRULE — rt.RRULE уже содержит итоговое правило (из периода или кастомное)
		if rt.RRULE == nil || *rt.RRULE == "" {
			return 0, fmt.Errorf("%w: no recurrence rule available", ErrInvalidInput)
		}

		rule, err := rrule.StrToRRule(*rt.RRULE)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid rrule: %v", ErrInvalidInput, err)
		}

		occurrences = rule.Between(fromUTC, toUTC, true)
	}

	// Фильтруем уже сгенерированные даты
	// var filtered []time.Time
	// if rt.LastGeneratedAt != nil {
	// 	lastUTC := rt.LastGeneratedAt.UTC()
	// 	for _, occ := range occurrences {
	// 		if occ.After(lastUTC) {
	// 			filtered = append(filtered, occ)
	// 		}
	// 	}
	// } else {
	// 	filtered = occurrences
	// }

	var filtered []time.Time
	if rt.LastGeneratedAt != nil {
		lastDate := rt.LastGeneratedAt.UTC().Truncate(24 * time.Hour)
		for _, occ := range occurrences {
			occDate := occ.UTC().Truncate(24 * time.Hour)
			if occDate.After(lastDate) {
				filtered = append(filtered, occ)
			}
		}
	} else {
		filtered = occurrences
	}

	// Создаём задачи
	generated := 0
	now := s.now()

	for _, occ := range filtered {
		titleWithDate := fmt.Sprintf("%s (%s)", rt.Title, occ.Format("02.01.2006"))
		task := &taskdomain.Task{
			Title:        titleWithDate,
			Description:  rt.Description,
			Status:       rt.Status,
			RepeatTaskID: &rt.ID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.taskRepo.Create(ctx, task); err != nil {
			continue
		}
		generated++
	}

	// Обновляем LastGeneratedAt по последней сгенерированной дате
	if len(filtered) > 0 {
		maxDate := filtered[len(filtered)-1]
		_ = s.repeatRepo.UpdateLastGeneratedAt(ctx, repeatTaskID, maxDate)
	}

	return generated, nil
}

// =============== Массовая генерация (на сегодня) ===============

func (s *Service) GenerateAllTasks(ctx context.Context) (int, error) {
	activeTasks, err := s.repeatRepo.ListActive(ctx)
	if err != nil {
		return 0, err
	}

	now := s.now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := from.Add(24*time.Hour - time.Second)

	total := 0
	for _, rt := range activeTasks {
		count, err := s.GenerateTasksForRepeatTask(ctx, rt.ID, from, to)
		if err != nil {
			continue
		}
		total += count
	}
	return total, nil
}

// =============== Валидация ===============

func validatePeriodInput(code, title string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("%w: code is required", ErrInvalidInput)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	return nil
}

// validateRepeatTaskInput проверяет что задано ровно одно из трёх правил повторения
func validateRepeatTaskInput(periodID *int64, rruleStr *string, customDates []time.Time) error {
	hasPeriod := periodID != nil
	hasRRULE := rruleStr != nil && *rruleStr != ""
	hasCustom := len(customDates) > 0

	// Несовместимые комбинации
	if hasPeriod && hasRRULE {
		return fmt.Errorf("%w: period_id and rrule cannot be used together, rrule is derived from period automatically", ErrInvalidInput)
	}
	if hasPeriod && hasCustom {
		return fmt.Errorf("%w: period_id and custom_dates cannot be used together", ErrInvalidInput)
	}
	if hasRRULE && hasCustom {
		return fmt.Errorf("%w: rrule and custom_dates cannot be used together", ErrInvalidInput)
	}

	// Должно быть задано хотя бы одно
	if !hasPeriod && !hasRRULE && !hasCustom {
		return fmt.Errorf("%w: one of period_id, rrule, or custom_dates must be provided", ErrInvalidInput)
	}

	return nil
}