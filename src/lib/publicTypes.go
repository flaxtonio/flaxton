package lib

import (
	"time"
)

type TransferContainerCall struct {
	Name 		string          `json:"name"`
	Cmd 		string 			`json:"cmd"`
	ImageName 	string			`json:"image_name"`
	ImageId 	string          `json:"image_id"`
	NeedToRun	bool        	`json:"need_to_run"`
	Authorization string 		`json:"authorization"`
}

// TaskStack types
const (
	TaskContainerTransfer = 1
)

type TaskType int

type Task struct  {
	TaskID string       `json:"task_id"`
	Type TaskType  		`json:"type"`
	Data interface{} 	`json:"data"`
	Cron bool 			`json:"cron"`
	StartTime time.Time `json:"start_time"`
	EndTime time.Time 	`json:"end_time"`
}
type TaskResult struct  {
	TaskID string       `json:"task_id"`
	StartTime time.Time `json:"start_time"`
	EndTime time.Time 	`json:"end_time"`
	Error bool          `json:"error"`
	Message string      `json:"message"`
}