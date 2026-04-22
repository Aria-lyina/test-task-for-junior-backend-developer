package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/usecase/repeat"
)

type RepeatTaskHandler struct {
	usecase repeat.Usecase
}

func NewRepeatTaskHandler(usecase repeat.Usecase) *RepeatTaskHandler {
	return &RepeatTaskHandler{usecase: usecase}
}

// =============== Periods ===============

func (h *RepeatTaskHandler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	var req periodMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.CreatePeriod(r.Context(), repeat.CreatePeriodInput{
		Code:          req.Code,
		Title:         req.Title,
		RRULETemplate: req.RRULETemplate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newPeriodDTO(created))
}

func (h *RepeatTaskHandler) GetPeriodByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	period, err := h.usecase.GetPeriodByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newPeriodDTO(period))
}


func (h *RepeatTaskHandler) UpdatePeriod(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req periodMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.UpdatePeriod(r.Context(), id, repeat.UpdatePeriodInput{
		Code:          req.Code,
		Title:         req.Title,
		RRULETemplate: req.RRULETemplate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newPeriodDTO(updated))
}

func (h *RepeatTaskHandler) DeletePeriod(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeletePeriod(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RepeatTaskHandler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	periods, err := h.usecase.ListPeriods(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]periodDTO, len(periods))
	for i, p := range periods {
		response[i] = newPeriodDTO(&p)
	}

	writeJSON(w, http.StatusOK, response)
}


func (h *RepeatTaskHandler) GenerateAllTasks(w http.ResponseWriter, r *http.Request) {
    
    count, err := h.usecase.GenerateAllTasks(r.Context())
    if err != nil {
        // h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate tasks: %v", err))
        return
    }
    
    response := GenerateTasksResponse{
        Message:        "tasks generated successfully",
        GeneratedCount: count,
    }
    
    writeJSON(w, http.StatusOK, response)
}


// =============== RepeatTasks ===============

func (h *RepeatTaskHandler) CreateRepeatTask(w http.ResponseWriter, r *http.Request) {
	var req repeatTaskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	customDates, err := parseDateStrings(req.CustomDates)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid custom_dates format, expected YYYY-MM-DD"))
		return
	}

	created, err := h.usecase.CreateRepeatTask(r.Context(), repeat.CreateRepeatTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      task.Status(req.Status),
		PeriodID:    req.PeriodID,
		RRULE:       req.RRULE,
		CustomDates: customDates,
		Enabled:     req.Enabled,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newRepeatTaskDTO(created))
}

func (h *RepeatTaskHandler) GetRepeatTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rt, err := h.usecase.GetRepeatTaskByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRepeatTaskDTO(rt))
}

func (h *RepeatTaskHandler) UpdateRepeatTask(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req repeatTaskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	customDates, err := parseDateStrings(req.CustomDates)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid custom_dates format, expected YYYY-MM-DD"))
		return
	}

	var status *task.Status
	if req.Status != "" {
		s := task.Status(req.Status)
		status = &s
	}

	updated, err := h.usecase.UpdateRepeatTask(r.Context(), id, repeat.UpdateRepeatTaskInput{
		Title:       &req.Title,
		Description: &req.Description,
		Status:      status,
		PeriodID:    req.PeriodID,
		RRULE:       req.RRULE,
		CustomDates: customDates,
		Enabled:     &req.Enabled,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRepeatTaskDTO(updated))
}

func (h *RepeatTaskHandler) DeleteRepeatTask(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeleteRepeatTask(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RepeatTaskHandler) ListActiveRepeatTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.ListActiveRepeatTasks(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]repeatTaskDTO, len(tasks))
	for i, rt := range tasks {
		response[i] = newRepeatTaskDTO(&rt)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *RepeatTaskHandler) ListRepeatTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.ListRepeatTasks(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]repeatTaskDTO, len(tasks))
	for i, rt := range tasks {
		response[i] = newRepeatTaskDTO(&rt)
	}
	writeJSON(w, http.StatusOK, response)
}



func (h *RepeatTaskHandler) GenerateTasks(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	from, err := parseDate(req.From)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid 'from' date format"))
		return
	}
	to, err := parseDate(req.To)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid 'to' date format"))
		return
	}

	count, err := h.usecase.GenerateTasksForRepeatTask(r.Context(), id, from, to)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"generated": count,
	})
}

func (h *RepeatTaskHandler) GetTasksWithRepeatInfo(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	infoList, err := h.usecase.GetTasksWithRepeatInfo(r.Context(), limit, offset)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskWithRepeatInfoDTO, len(infoList))
	for i, info := range infoList {
		response[i] = newTaskWithRepeatInfoDTO(info)
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateTaskWithRepeat создаёт шаблон повторения и первую задачу от него за один запрос
func (h *RepeatTaskHandler) CreateTaskWithRepeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Task   taskMutationDTO        `json:"task"`
		Repeat repeatTaskMutationDTO  `json:"repeat"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	customDates, err := parseDateStrings(req.Repeat.CustomDates)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid custom_dates format, expected YYYY-MM-DD"))
		return
	}

	createdTask, createdRepeat, err := h.usecase.CreateTaskWithRepeat(
		r.Context(),
		repeat.CreateTaskInput{
			Title:       req.Task.Title,
			Description: req.Task.Description,
			Status:      req.Task.Status,
		},
		repeat.CreateRepeatTaskInput{
			Title:       req.Repeat.Title,
			Description: req.Repeat.Description,
			Status:      task.Status(req.Repeat.Status),
			PeriodID:    req.Repeat.PeriodID,
			RRULE:       req.Repeat.RRULE,
			CustomDates: customDates,
			Enabled:     req.Repeat.Enabled,
		},
	)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"task":        newTaskDTO(createdTask),
		"repeat_task": newRepeatTaskDTO(createdRepeat),
	})
}

// =============== DTO ===============

type periodMutationDTO struct {
	Code          string  `json:"code"`
	Title         string  `json:"title"`
	RRULETemplate *string `json:"rrule_template,omitempty"`
}

type periodDTO struct {
	ID            int64   `json:"id"`
	Code          string  `json:"code"`
	Title         string  `json:"title"`
	RRULETemplate *string `json:"rrule_template,omitempty"`
}

func newPeriodDTO(p *task.Period) periodDTO {
	return periodDTO{
		ID:            p.ID,
		Code:          p.Code,
		Title:         p.Title,
		RRULETemplate: p.RRULETemplate,
	}
}

type repeatTaskMutationDTO struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	PeriodID    *int64   `json:"period_id,omitempty"`
	RRULE       *string  `json:"rrule,omitempty"`
	CustomDates []string `json:"custom_dates,omitempty"`
	Enabled     bool     `json:"enabled"`
}

type repeatTaskDTO struct {
	ID              int64      `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	PeriodID        *int64     `json:"period_id,omitempty"`
	RRULE           *string    `json:"rrule,omitempty"`
	CustomDates     []string   `json:"custom_dates,omitempty"`
	Enabled         bool       `json:"enabled"`
	LastGeneratedAt *time.Time `json:"last_generated_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func newRepeatTaskDTO(rt *task.RepeatTask) repeatTaskDTO {
	// Конвертируем DateArray ([]time.Time) в []string для JSON ответа
	customDates := make([]string, len(rt.CustomDates))
	for i, d := range rt.CustomDates {
		customDates[i] = d.Format("2006-01-02")
	}

	return repeatTaskDTO{
		ID:              rt.ID,
		Title:           rt.Title,
		Description:     rt.Description,
		Status:          string(rt.Status),
		PeriodID:        rt.PeriodID,
		RRULE:           rt.RRULE,
		CustomDates:     customDates,
		Enabled:         rt.Enabled,
		LastGeneratedAt: rt.LastGeneratedAt,
		CreatedAt:       rt.CreatedAt,
		UpdatedAt:       rt.UpdatedAt,
	}
}

type taskWithRepeatInfoDTO struct {
	Task       taskDTO        `json:"task"`
	RepeatTask *repeatTaskDTO `json:"repeat_task,omitempty"`
	Period     *periodDTO     `json:"period,omitempty"`
}

func newTaskWithRepeatInfoDTO(info task.TaskWithRepeatInfo) taskWithRepeatInfoDTO {
	dto := taskWithRepeatInfoDTO{
		Task: newTaskDTO(&info.Task),
	}
	if info.RepeatTask != nil {
		rtDTO := newRepeatTaskDTO(info.RepeatTask)
		dto.RepeatTask = &rtDTO
	}
	if info.Period != nil {
		pDTO := newPeriodDTO(info.Period)
		dto.Period = &pDTO
	}
	return dto
}

// =============== Вспомогательные функции ===============

func parseDateStrings(dates []string) ([]time.Time, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	result := make([]time.Time, len(dates))
	for i, ds := range dates {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return nil, err
		}
		result[i] = t
	}
	return result, nil
}

func parseDate(s string) (time.Time, error) {
	// Сначала пробуем YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Затем RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid date format, expected YYYY-MM-DD or RFC3339")
}

type GenerateTasksResponse struct {
    Message        string `json:"message"`
    GeneratedCount int    `json:"generated_count"`
}