package handler

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"github.com/osmso/clock/models"
	"github.com/osmso/clock/common"
	"github.com/osmso/clock/database"
	"time"
	"github.com/gorilla/mux"
	"strconv"
	"math/rand"
	"fmt"
)

var God = make(map[string]*common.ClocksStore)

func initClocks(r *http.Request) *common.ClocksStore {
	vars := mux.Vars(r)
	clock := vars["timer"]

	return InitNewClocks(clock)
}

func InitNewClocks(clock string) *common.ClocksStore {
	_, ok := God[clock]
	if ok {
	} else {
		God[clock] = common.NewClocksStore()
	}

	return God[clock]
}

func InitDbClocks(clocks *([]models.ClockExt)) {
	for _, clock := range *clocks {
		x := 1000*500 + rand.Intn(1000*1000*5)
		time.Sleep(time.Duration(x))
		JobsQueue := MakerQueueMiddle(clock.Timer)
		Timer := InitNewClocks(clock.Timer)

		content := string(clock.Content.([]byte))
		poptimes := clock.PopTimes
		Timer.AddClock(clock.Tid, clock.Source, time.Second*time.Duration(clock.Interval), clock.Repeat, func() {
			JobsQueue.Push(content, poptimes)
		})
	}
}

func NewClock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timer := vars["timer"]

	JobsQueue := MakerQueue(r)
	Timer := initClocks(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	clock := new(models.Clock)
	err = json.Unmarshal(body, clock)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	content, err := json.Marshal(clock.Content)
	fmt.Println(string(content))
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	store, err := json.Marshal(clock)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	success, err, id := Timer.AddClock(clock.Tid, string(store), time.Second*time.Duration(clock.Interval), clock.Repeat, func() {
		JobsQueue.Push(string(content), clock.PopTimes)
	})

	if err != nil {
		ErrorResponse(w, err)
		return
	}

	if success && common.AppConfig.UseDB {
		clockExt := new(models.ClockExt)
		err = json.Unmarshal(body, clockExt)
		if err != nil {

		} else {
			clockExt.Tid = id
			clockExt.Timer = timer
			clockExt.Content = string(content)
			clockExt.Delete = false
			clockExt.Source = string(store)
			database.CreateDbClock(clockExt)
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func UpdateClock(w http.ResponseWriter, r *http.Request) {
	JobsQueue := MakerQueue(r)
	vars := mux.Vars(r)
	timer := vars["timer"]

	_, ok := God[timer]
	if ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	clock := new(models.ClockExt)
	err = json.Unmarshal(body, clock)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	God[timer].RemoveClock(clock.Tid)
	database.DeleteDbClock(clock)
	content, err := json.Marshal(clock.Content)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	store, err := json.Marshal(clock)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	success, err, id := God[timer].AddClock(clock.Tid, string(store), time.Second*time.Duration(clock.Interval), clock.Repeat, func() {
		JobsQueue.Push(string(content), clock.PopTimes)
	})

	if err != nil {
		ErrorResponse(w, err)
		return
	}

	if success && common.AppConfig.UseDB {
		clockExt := new(models.ClockExt)
		err = json.Unmarshal(body, clockExt)
		if err != nil {

		} else {
			clockExt.Tid = id
			clockExt.Timer = timer
			clockExt.Content = string(content)
			clockExt.Delete = false
			database.CreateDbClock(clockExt)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func GetClock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timer := vars["timer"]

	vals := r.URL.Query()

	begin_var := vals.Get("begin")
	end_var := vals.Get("end")

	if len(begin_var) == 0 || len(end_var) == 0 {
		begin_var = "0"
		end_var = "30"
	}

	begin, err := strconv.Atoi(begin_var)
	if err != nil {
		begin = 0
	}

	end, err := strconv.Atoi(end_var)
	if err != nil {
		end = 30
	}

	_, ok := God[timer]
	if ok {

	} else {
		w.WriteHeader(http.StatusCreated)
		return
	}

	craws := God[timer].GetClocks(begin, end)

	bytes, err := json.Marshal(map[string]interface{}{"timers": craws})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeJsonResponse(w, bytes)
}

func DeleteClock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timer := vars["timer"]

	_, ok := God[timer]
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	clock := new(models.ClockExt)
	err = json.Unmarshal(body, clock)
	if err != nil {
		ErrorResponse(w, err)
		return
	}

	God[timer].RemoveClock(clock.Tid)
	clock.Delete = true
	if common.AppConfig.UseDB {
		database.DeleteDbClock(clock)
	}
	w.WriteHeader(http.StatusNoContent)
}

func CleanTimer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timer := vars["timer"]

	_, ok := God[timer]
	if ok {
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	God[timer].CleanInsectGod()
	delete(God, timer)
	w.WriteHeader(http.StatusNoContent)
}

func GetTimers(w http.ResponseWriter, r *http.Request) {

	Timers := make(map[string][]string)
	for timer := range God {
		Timers["timers"] = append(Timers["timers"], timer)
	}

	bytes, err := json.Marshal(Timers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeJsonResponse(w, bytes)
}

func ClocksCounts(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	timer := vars["timer"]

	_, ok := God[timer]
	if ok {
		bytes, err := json.Marshal(map[string]interface{}{"counts": God[timer].GetClockNum()})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		writeJsonResponse(w, bytes)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func writeJsonResponse(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bytes)
}

func ErrorResponse(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	w.WriteHeader(http.StatusExpectationFailed)
}
