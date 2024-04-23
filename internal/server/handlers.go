package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const tenMB = 10 << 20

type registerResp struct {
	BiometricID []float64 `json:"biometric_id"`
}

type matchResp struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(tenMB); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error retrieving the image file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error reading the image file", http.StatusInternalServerError)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Failed to get current working directory", http.StatusInternalServerError)
		return
	}

	tempDirPath := filepath.Join(cwd, "temp-images")

	tempFile, err := os.CreateTemp(tempDirPath, "upload-*.jpeg")
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error creating a temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(fileBytes); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error writing to temporary file", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("python3", "generate_biometric_id.py", tempFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error executing Python script", http.StatusInternalServerError)
		return
	}

	var biometricID []float64
	if err := json.Unmarshal(output, &biometricID); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error unmarshalling python output", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if err := a.db.CreateUser(biometricID, email); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error inserting user to db", http.StatusInternalServerError)
		return
	}

	if err := tempFile.Close(); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Err closing temp file", http.StatusInternalServerError)
		return
	}

	if err := a.db.ExportEmbeddings(); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error exporting embeddings from database", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(registerResp{BiometricID: biometricID})
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Err marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (a *API) MatchHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(tenMB); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error retrieving the image file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error reading the image file", http.StatusInternalServerError)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Failed to get current working directory", http.StatusInternalServerError)
		return
	}

	tempDirPath := filepath.Join(cwd, "temp-images")

	tempFile, err := os.CreateTemp(tempDirPath, "upload-*.jpeg")
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error creating a temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(fileBytes); err != nil {
		fmt.Println("err: ", err)
		http.Error(w, "Error writing to temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	galleryEmbeddingsPath := filepath.Join(cwd, "gallery_embeddings.npy")

	cmd := exec.Command("python3", "find_match_in_gallery.py", tempFile.Name(), galleryEmbeddingsPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("output: ", string(output))
		fmt.Println("err: ", err)
		http.Error(w, "Error executing Python script", http.StatusInternalServerError)
		return
	}

	userIDStr := strings.TrimSpace(string(output))
	if userIDStr == "" || userIDStr == "None" {
		fmt.Println("err: ", err)
		http.Error(w, "No match found", http.StatusNotFound)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		fmt.Println("err:", err)
		http.Error(w, "Error processing match result", http.StatusInternalServerError)
		return
	}

	user, err := a.db.GetUserByID(userID)
	if err != nil {
		fmt.Println("err:", err)
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		return
	}

	resp := matchResp{
		UserID: user.UserID,
		Email:  user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("err:", err)
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}

	w.Write(respBytes)
}
