package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and creates the necessary tables if they do not exist
func InitDB() {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		firstname TEXT NOT NULL,
		lastname TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		token TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	// Ensure the time_card_details table is created
	CreateTimeCardDetailsTable()

}

// CreateTimeCardDetailsTable creates the table for time card details if it does not exist
func CreateTimeCardDetailsTable() {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS time_card_details (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pay_period TEXT NOT NULL,
		year TEXT NOT NULL,
		campaign_id INTEGER NOT NULL,
		firm_id INTEGER NOT NULL,
		days TEXT NOT NULL,
		regular REAL NOT NULL,
		holidays REAL NOT NULL,
		vacation REAL NOT NULL,
		date TEXT NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

// AgentExists checks if an agent with the given information already exists in the database
func AgentExists(agent Agent) bool {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM agents WHERE email=? OR username=?)`
	err = db.QueryRow(query, agent.Email, agent.Username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	return exists
}

// InsertAgent inserts a new agent into the database
func InsertAgent(agent Agent) error {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		return err
	}
	defer db.Close()

	insertSQL := `INSERT INTO agents (id, firstname, lastname, email, username, password) VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(agent.ID, agent.FirstName, agent.LastName, agent.Email, agent.Username, agent.Password)
	if err != nil {
		return err
	}

	return nil
}

// InsertTimeCardDetail inserts a new time card detail into the database
func InsertTimeCardDetail(payPeriod, year string, detail TimeCardDetail) error {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		return err
	}
	defer db.Close()

	insertSQL := `INSERT INTO time_card_details (pay_period, year, campaign_id, firm_id, days, regular, holidays, vacation, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(payPeriod, year, detail.CampaignId, detail.FirmId, detail.Days, detail.Regular, detail.Holidays, detail.Vacation, detail.Date)
	if err != nil {
		return err
	}

	return nil
}

// GetAllTimeCardDetails retrieves all time card details from the database
func GetAllTimeCardDetails() ([]Schedule, error) {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT pay_period, year, campaign_id, firm_id, days, regular, holidays, vacation, date FROM time_card_details ORDER BY pay_period, year`
	//query = `SELECT id, campaign_id, firm_id, days, regular, holidays, vacation, date FROM time_card_details`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying time card details: %v", err)
		return nil, fmt.Errorf("error querying time card details: %w", err)
	}
	defer rows.Close()

	var schedules []Schedule
	var currentSchedule *Schedule
	for rows.Next() {
		var detail TimeCardDetail
		var payPeriod, year string
		err := rows.Scan(&payPeriod, &year, &detail.CampaignId, &detail.FirmId, &detail.Days, &detail.Regular, &detail.Holidays, &detail.Vacation, &detail.Date)
		if err != nil {
			return nil, err
		}
		if currentSchedule == nil || currentSchedule.PayPeriod != payPeriod || currentSchedule.Year != year {
			if currentSchedule != nil {
				schedules = append(schedules, *currentSchedule)
			}
			currentSchedule = &Schedule{
				PayPeriod:       payPeriod,
				Year:            year,
				TimeCardDetails: []TimeCardDetail{detail},
			}
		} else {
			currentSchedule.TimeCardDetails = append(currentSchedule.TimeCardDetails, detail)
		}
	}
	if currentSchedule != nil {
		schedules = append(schedules, *currentSchedule)
	}
	return schedules, nil
}

// GetTimeCardDetailById retrieves a specific time card detail by ID from the database
func GetTimeCardDetailById(id string) (*TimeCardDetail, error) {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, campaign_id, firm_id, days, regular, holidays, vacation, date FROM time_card_details WHERE id = ?`
	var detail TimeCardDetail
	//err = db.QueryRow(query, id).Scan(&detail.ID, &detail.CampaignId, &detail.FirmId, &detail.Days, &detail.Regular, &detail.Holidays, &detail.Vacation, &detail.Date)
	err = db.QueryRow(query, id).Scan(&detail.ID, &detail.CampaignId, &detail.FirmId, &detail.Days, &detail.Regular, &detail.Holidays, &detail.Vacation, &detail.Date)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// GetAllAgentsWithTokens retrieves all agents with their JWT tokens from the database
func GetAllAgentsWithTokens() ([]Agent, error) {
	db, err := sql.Open("sqlite3", "agent.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, firstname, lastname, email, username, token FROM agents`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying agents with tokens: %v", err)
		return nil, fmt.Errorf("error querying agents with tokens: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var token sql.NullString
		err := rows.Scan(&agent.ID, &agent.FirstName, &agent.LastName, &agent.Email, &agent.Username, &token)
		if err != nil {
			return nil, err // Return the error if scanning fails
		}
		if token.Valid {
			agent.Token = token.String
			// Calculate the age of the token
			tokenAge, ageErr := calculateTokenAge(token.String)
			if ageErr != nil {
				log.Printf("Error calculating token age: %v", ageErr)
				// If there's an error calculating the token age, we can decide to continue without token age or return an error
				// For now, we'll continue without token age
			} else {
				agent.TokenAge = tokenAge // Set the token age if it was calculated successfully
			}
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// calculateTokenAge calculates the age of the token based on the expiration time
func calculateTokenAge(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your_secret_key"), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			duration := time.Since(expirationTime)
			return fmt.Sprintf("%v ago", duration), nil
		}
	}

	return "", fmt.Errorf("unable to calculate token age")
}
