package models

type Task struct {
	ID	       int        `json:"id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
}

type IncomingTask struct {
	TaskText   string	  `json:"tasktext"`
}

type UpdateTaskInput struct {
	Status 	   string     `json:"status"`
}
