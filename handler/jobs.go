package handler

import (
	. "github.com/osmso/clock/common"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"github.com/osmso/clock/models"
	"strconv"
	"github.com/gorilla/mux"
	"errors"
	"log"
)

var JobsQueueOfClocks = make(map[string]*JobQueue)

type Job struct {
	Id      string
	Content interface{}
}

func InitQueue(r *http.Request) (*JobQueue, error) {
	vars := mux.Vars(r)
	clock := vars["timer"]

	_, ok := JobsQueueOfClocks[clock]
	if ok {
	} else {
		return nil, errors.New("no found this jobs queue of this timer")
	}

	return JobsQueueOfClocks[clock], nil
}

func MakerQueue(r *http.Request) (*JobQueue) {
	vars := mux.Vars(r)
	clock := vars["timer"]

	return MakerQueueMiddle(clock)
}

func MakerQueueMiddle(clock string) (*JobQueue) {
	_, ok := JobsQueueOfClocks[clock]
	if ok {
	} else {
		JobsQueueOfClocks[clock] = NewJobQueue()
	}

	return JobsQueueOfClocks[clock]
}

func NewJobs(w http.ResponseWriter, r *http.Request) {
	JobsQueue := MakerQueue(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("NewJobs Error",err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}

	jobCore := new(models.JobCore)
	err = json.Unmarshal(body, jobCore)
	if err != nil {
		log.Println("NewJobs Error",err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}

	content, err := json.Marshal(jobCore.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusExpectationFailed)
	} else {
		JobsQueue.Push(string(content), jobCore.PopTimes)
		w.WriteHeader(http.StatusCreated)
	}
}

func GetTimerJobs(w http.ResponseWriter, r *http.Request) {
	JobsQueue, err := InitQueue(r)

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	params := r.URL.Query()
	num, err := strconv.Atoi(params.Get("num"))

	lens := JobsQueue.Length()

	if err != nil {
		num = 1
	} else if num > 1024 {
		num = 1024
	}

	if num > 5*lens {
		num = 5 * lens
	}

	jobs := []interface{}{}
	var count = 0
	for i := 0; i < num; i++ {
		job := JobsQueue.Pop()
		if job != nil {
			jobs = append(jobs, job)
			count++
		} else {
			break
		}
	}

	bytes, err := json.Marshal(map[string]interface{}{"jobs": jobs, "count": count})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeJsonResponse(w, bytes)
}

func DeleteTimerJob(w http.ResponseWriter, r *http.Request) {
	JobsQueue, err := InitQueue(r)

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	Job := new(models.Job)
	err = json.Unmarshal(body, Job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	JobsQueue.Remove(Job.Id)
	w.WriteHeader(http.StatusNoContent)
}

func CleanJob(w http.ResponseWriter, r *http.Request) {
	JobsQueue, err := InitQueue(r)

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	JobsQueue.Clean()
	w.WriteHeader(http.StatusNoContent)
}

func TimerJobsLength(w http.ResponseWriter, r *http.Request) {
	JobsQueue, err := InitQueue(r)

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	values := map[string]int{"length": JobsQueue.Length()}

	jsonValue, err := json.Marshal(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeJsonResponse(w, jsonValue)
}
