package task
 
import "time"
 
type Status string
 
const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)
 
type Task struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       Status    `json:"status"`
	RepeatTaskID *int64    `json:"repeat_task_id,omitempty"` // NULL если задача создана вручную, иначе ссылка на repeat_tasks.id
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
 
func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
 
