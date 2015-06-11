package lib

import (
	"encoding/json"
)

// TaskStack types
const (
	TaskImageTransfer = "image_transfer"
	TaskSetDaemonName = "set_daemon_name"
	TaskAddChildServer = "add_child_server"
	TaskAddBalancingImage = "add_balancing_image"
	TaskStartBalancerPort = "start_balancing_port"
	TaskStopBalancerPort = "stop_balancing_port"
	TaskStartContainer = "start_container"
	TaskStopContainer = "stop_container"
	TaskPauseContainer = "pause_container"
	TaskCreateContainer = "create_container"
)

type TaskType string

type Task struct  {
	TaskID string       `json:"task_id"`
	Type TaskType  		`json:"type"`
	Data interface{} 	`json:"data"`
	Cron bool 			`json:"cron"`
	StartTime int64 	`json:"start_time,omitempty"`
	EndTime int64 		`json:"end_time,omitempty"`
}
type TaskResult struct  {
	TaskID string       `json:"task_id"`
	StartTime int64 	`json:"start_time,omitempty"`
	EndTime int64 		`json:"end_time,omitempty"`
	Error bool          `json:"error"`
	Done bool           `json:"done"`
	Message string      `json:"message"`
}

type TaskSendResponse struct {
	Error bool 		`json:"error"`
	Message string 	`json:"message"`
	TaskId string   `json:"task_id"`
}

func (t *Task) ConvertData(obj interface{}) error {
	b, err := json.Marshal(t.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}
	return nil
}