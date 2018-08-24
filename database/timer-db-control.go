package database

import (
	"github.com/osmso/clock/models"
)

func GetDbClocks(clocks *([]models.ClockExt)) *([]models.ClockExt) {
	var getDb = GetDb()
	if err := getDb.Find(&clocks).Error; err != nil {
	}

	return clocks
}

func CreateDbClock(clock *models.ClockExt) {
	var getDb = GetDb()
	getDb.Create(&clock)
}

func DeleteDbClock(clock *models.ClockExt) {
	var getDb = GetDb()
	getDb.Where(clock.Tid).Delete(&clock)
}