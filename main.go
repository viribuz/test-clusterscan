package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/AgentRegistration", AgentRegistrationHandler).Methods("POST")
	router.HandleFunc("/SubmitLogin", SubmitLoginHandler).Methods("POST")

	router.HandleFunc("/ScheduleSubmission", ScheduleSubmissionHandler).Methods("POST")
	router.HandleFunc("/GetSchedules", GetSchedulesHandler).Methods("GET")
	router.HandleFunc("/GetScheduleById/{id}", GetScheduleByIdHandler).Methods("GET")
	router.HandleFunc("/RenewToken", RenewTokenHandler).Methods("POST")
	router.HandleFunc("/GetAgentsWithTokens", GetAgentsWithTokensHandler).Methods("GET")

	// err := InitDB()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize database: %v", err)
	// }
	InitDB()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
