package handlers

import (
	"errors"
	"net/http"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/usecase/repeat"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	// Общие ошибки
	case errors.Is(err, taskdomain.ErrNotFound),
		errors.Is(err, taskdomain.ErrPeriodNotFound),
		errors.Is(err, repeat.ErrRepeatTaskNotFound):
		writeError(w, http.StatusNotFound, err)

	// Ошибки валидации
	case errors.Is(err, taskusecase.ErrInvalidInput),
		errors.Is(err, repeat.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)

	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}