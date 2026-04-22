package task

import (       // импорты идут сразу после package
    "errors"
    // "fmt"
)

var (
    ErrPeriodNotFound = errors.New("period not found")
    ErrNotFound   = errors.New("task not found")
    ErrRepeatTaskNotFound = errors.New("repeat task not found")
)
