package models

type User struct {
	ID			  int	 `json:"id"`
	Email   	  string `json:"email"`
	Passwordhash  string `json:"-"`//"-" означает никогда не выводить это поле в джсон ответах
}

type RegisterInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}