package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description" gorm:"type:varchar(1000)"`
	Price       float64 `json:"price" gorm:"not null;type:decimal(10,2)"`
	Quantity    int     `json:"quantity" gorm:"not null;default:0"`
}
