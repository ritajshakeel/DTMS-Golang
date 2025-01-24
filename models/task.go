package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	AssignedTo       *uint     `json:"assigned_to"`
	User             *User     `json:"user" gorm:"foreignKey:AssignedTo"`
	PlannedStartTime time.Time `json:"planned_start_time"`
	PlannedEndTime   time.Time `json:"planned_end_time"`
	ActualStartTime  time.Time `json:"actual_start_time"`
	ActualEndTime    time.Time `json:"actual_end_time"`
	Seconds          int64     `json:"seconds"`
}
