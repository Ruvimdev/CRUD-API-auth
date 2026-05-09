package models

type Task struct {
	ID	       int        `json:"id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
}

type IncomingTask struct {
	TaskText   string	  `json:"tasktext" validate:"required,min=3"`
}

type UpdateTaskInput struct {
											//oneof принимающие только эти слова, остальное ошибка
	Status 	   string     `json:"status" validate:"required,oneof=active done"`
}
