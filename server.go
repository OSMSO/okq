package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/osmso/clock/handler"
	"github.com/osmso/clock/common"
	"github.com/osmso/clock/database"
	"log"
)

func main() {
	common.InitConfig()
	if common.AppConfig.UseDB {
		go func() {
			log.Println("UseDB Init")
			clocks := database.Init()
			handler.InitDbClocks(clocks)
			log.Println("DB clock Init Finish")
		}()
	}

	router := mux.NewRouter().StrictSlash(true)
	sub := router.PathPrefix("/api/v1/timers").Subrouter()

	// Timers Api
	sub.Methods("GET").Path("/").HandlerFunc(handler.GetTimers)
	sub.Methods("GET").Path("/{timer}").HandlerFunc(handler.GetClock)
	sub.Methods("GET").Path("/{timer}/counts").HandlerFunc(handler.ClocksCounts)
	sub.Methods("PUT").Path("/{timer}").HandlerFunc(handler.UpdateClock)
	sub.Methods("POST").Path("/{timer}").HandlerFunc(handler.NewClock)
	sub.Methods("DELETE").Path("/{timer}").HandlerFunc(handler.DeleteClock)
	sub.Methods("DELETE").Path("/{timer}/all").HandlerFunc(handler.CleanTimer)

	http.ListenAndServe(":3000", router)
}
