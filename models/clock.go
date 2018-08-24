package models

import "github.com/jinzhu/gorm"

type Model struct {
	ID uint64 `gorm:"primary_key" json:"id"`
}

type Clock struct {
	Tid      string      `json:"tid"`
	Repeat   uint64      `json:"repeat"`
	PopTimes int64       `json:"poptimes"`
	Interval int64       `json:"interval"`
	Content  interface{} `gorm:"type:LONGTEXT" json:"content"`
}

type ClockExt struct {
	gorm.Model
	Clock
	Timer  string `json:"timer"`
	Delete bool   `json:"deleted"`
	Source string `json:"source"`
}
