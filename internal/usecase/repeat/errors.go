package repeat
 
import "errors"
 
var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrRepeatTaskNotFound = errors.New("repeat task not found")
)
 