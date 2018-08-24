package models

type Job struct {
	Id int64 `json:"id"`
}

type JobCore struct {
	Id       string      `json:"id"`
	PopTimes int64       `json:"poptimes"`
	Content  interface{} `json:"Content"`
}