package models




type Category struct {
	ID		int  `json:"id"`
	Name	string	`json:"category"`
}

type CategoryInput struct {
	CategoryName	string	`json:"category_name", validate:"required,min=2"`
}