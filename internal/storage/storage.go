package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lib/pq"
)

type Db struct {
	*sql.DB
}

type Database interface {
	CreateUser(biometricID []float64, email string) error
	GetAllEmbeddings() ([]UserEmbedding, error)
	GetUserByID(id int) (UserEmbedding, error)
	ExportEmbeddings() error
}

type UserEmbedding struct {
	UserID    int       `json:"user_id"`
	Embedding []float64 `json:"embedding"`
	Email     string    `json:"email,omitempty"`
}

// NewDB creates a connection with our postgres database and returns it, otherwise an error.
func NewDB() (*Db, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("FACE_REC_SYSTEM_DB_HOST"),
		os.Getenv("FACE_REC_SYSTEM_DB_PORT"),
		os.Getenv("FACE_REC_SYSTEM_DB_USER"),
		os.Getenv("FACE_REC_SYSTEM_DB_PASSWORD"),
		os.Getenv("FACE_REC_SYSTEM_DB_NAME"),
		os.Getenv("FACE_REC_SYSTEM_SSL_MODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Db{db}, nil
}

// CreateUser inserts a new user with their biometric ID into the database.
func (db *Db) CreateUser(biometricID []float64, email string) error {
	// Convert the []float64 to a PostgreSQL ARRAY representation
	// PostgreSQL array literals are enclosed in braces and separated by commas.
	biometricStr := make([]string, len(biometricID))
	for i, v := range biometricID {
		biometricStr[i] = fmt.Sprintf("%f", v)
	}
	biometricArray := "{" + strings.Join(biometricStr, ",") + "}"

	stmt, err := db.Prepare("INSERT INTO users (biometric_id, email) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(biometricArray, email); err != nil {
		return err
	}

	return nil
}

// GetAllEmbeddings retrieves all biometric embeddings from the database.
func (db *Db) GetAllEmbeddings() ([]UserEmbedding, error) {
	var embeddings []UserEmbedding
	rows, err := db.Query("SELECT id, biometric_id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ue UserEmbedding
		if err := rows.Scan(&ue.UserID, pq.Array(&ue.Embedding)); err != nil {
			return nil, err
		}
		embeddings = append(embeddings, ue)
	}

	return embeddings, nil
}

// GetUserEmbeddingByID retrieves a user and their biometric_id embedding.
func (db *Db) GetUserByID(id int) (UserEmbedding, error) {
	query := "SELECT id, biometric_id, email FROM users WHERE id = $1"
	row := db.QueryRow(query, id)

	var ue UserEmbedding
	err := row.Scan(&ue.UserID, pq.Array(&ue.Embedding), &ue.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserEmbedding{}, fmt.Errorf("user with ID %d not found", id)
		}
		return UserEmbedding{}, fmt.Errorf("error retrieving user embedding by ID: %v", err)
	}

	return ue, nil
}

// ExportEmbeddings takes all the embeddings currently in today's gallery
// and writes them locally to a file for consumption by python scripts.
func (db *Db) ExportEmbeddings() error {
	embeddings, err := db.GetAllEmbeddings()
	if err != nil {
		return fmt.Errorf("fetching embeddings: %w", err)
	}

	jsonData, err := json.Marshal(embeddings)
	if err != nil {
		log.Printf("Error marshaling embeddings: %v\n", err)
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("err: ", err)
		return err
	}

	tempDirPath := filepath.Join(cwd, "temp-embeddings")

	tmpFile, err := os.CreateTemp(tempDirPath, "embeddings-*.json")
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(jsonData); err != nil {
		return fmt.Errorf("writing JSON to temporary file: %w", err)
	}
	defer tmpFile.Close()

	cmd := exec.Command("python3", "save_embeddings.py", tmpFile.Name(), "gallery_embeddings.npy")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("saving embeddings to .npy: %s, %w", string(output), err)
	}

	return nil
}
