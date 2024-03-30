package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)
import ( 
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// AgentRegistrationHandler handles the registration of a new agent
func AgentRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var agent Agent
	err := json.NewDecoder(r.Body).Decode(&agent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if any of the fields already exist in the database
	if AgentExists(agent) {
		http.Error(w, "Agent with the given information already exists", http.StatusConflict)
		return
	}

	// Generate a new JWT token for the agent ID
	agent.ID, err = GenerateJWT(agent.Username) // Use the username as a unique identifier for the token
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Insert the new agent into the database
	if err := InsertAgent(agent); err != nil {
		http.Error(w, "Failed to register agent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

// SubmitLoginHandler handles the login submission and token generation
func SubmitLoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := AuthenticateAgent(credentials.Username, credentials.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// ScheduleSubmissionHandler handles the submission of time card details
func ScheduleSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	var submission ScheduleSubmission
	err := json.NewDecoder(r.Body).Decode(&submission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	for i := range submission.TimeCardDetails {
		detail := &submission.TimeCardDetails[i]
		validDays := map[string]bool{"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true, "Friday": true, "Saturday": true, "Sunday": true}
		if _, ok := validDays[detail.Days]; !ok {
			http.Error(w, "Invalid day. Days must be one of Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, or Sunday.", http.StatusBadRequest)
			return
		}
		// Generate a UUID for each TimeCardDetail
		// Assuming ID should be an integer, we should not assign a UUID string to it.
		// If ID is meant to be a string, the TimeCardDetail struct in models.go should be updated accordingly.
		// For now, we will remove this line to prevent the error.
		// detail.ID = uuid.New().String()
		err := InsertTimeCardDetail(submission.PayPeriod, submission.Year, *detail)
		if err != nil {
			http.Error(w, "Failed to insert time card detail", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submission)
}

// VerifyAgentCredentials checks the provided credentials against the database
func VerifyAgentCredentials(username, password string) (string, bool) {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var storedPassword, agentID string
	err = db.QueryRow("SELECT id, password FROM agents WHERE username = ?", username).Scan(&agentID, &storedPassword)
	if err != nil || storedPassword != password {
		return "", false
	}
	return agentID, true
}

// GetSchedulesHandler retrieves all time card details
func GetSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	schedules, err := GetAllTimeCardDetails()
	if err != nil {
		http.Error(w, "Failed to retrieve schedules", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedules)
}

// GetScheduleByIdHandler retrieves a specific time card detail by ID
func GetScheduleByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	schedule, err := GetTimeCardDetailById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve schedule", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// RenewTokenHandler handles the renewal of JWT tokens
func RenewTokenHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"` // Assuming the token is passed in the request body
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims := &jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(request.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your_secret_key"), nil // Replace "your_secret_key" with the actual key used for signing JWT tokens
	})

	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	if iat, ok := (*claims)["iat"].(float64); ok { // Assuming "iat" (issued at) claim is used in the JWT token
		issuedAt := time.Unix(int64(iat), 0)
		if time.Since(issuedAt) > 24*time.Hour {
			newToken, err := RenewJWT(request.Token)
			if err != nil {
				http.Error(w, "Failed to renew token", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"token": newToken}) // Send the new token in the response
			return
		}
	}

	// If the token is not expired, inform the client
	http.Error(w, "Token is not yet expired", http.StatusBadRequest)
}

// GetAgentsWithTokensHandler handles the request for users and their JWT tokens
func GetAgentsWithTokensHandler(w http.ResponseWriter, r *http.Request) {
	agents, err := GetAllAgentsWithTokens()
	if err != nil {
		log.Printf("Error retrieving agents with tokens: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve agents with tokens: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(agents)
}
