package mental_rotation

import (
	"embed"
	_ "embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed mental_rotation.html
var frontendFile []byte

//go:embed images/*
var images embed.FS

type Task struct {
	ID            int       `json:"id"`
	Image         string    `json:"image"`
	CorrectAnswer bool      `json:"correctAnswer"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
}

type Result struct {
	ParticipantID string        `json:"participantId"`
	Image         string        `json:"image"`
	IsCorrect     bool          `json:"isCorrect"`
	TimeTaken     time.Duration `json:"timeTaken"`
	Timestamp     string        `json:"timestamp"`
}

var (
	tasks       []Task
	results     []Result
	mu          sync.RWMutex
	resultsFile string
)

func Init() {
	// Set up results file path
	resultsFile = filepath.Join("data", "mental_rotation_results.json")

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		panic(err)
	}

	// Load existing results if any
	if data, err := os.ReadFile(resultsFile); err == nil {
		if err := json.Unmarshal(data, &results); err != nil {
			panic(err)
		}
	}

	// Discover all JPG images in the images directory
	var imageFiles []string
	fs.WalkDir(images, "images", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".jpg") {
			imageFiles = append(imageFiles, d.Name())
		}
		return nil
	})

	// Sort image files to ensure consistent ordering
	sort.Strings(imageFiles)

	// Create tasks from discovered images
	tasks = make([]Task, len(imageFiles))
	for i, image := range imageFiles {
		correctAnswer := !strings.HasSuffix(strings.ToLower(image), "r.jpg")

		tasks[i] = Task{
			ID:            i + 1,
			Image:         image,
			CorrectAnswer: correctAnswer,
		}
	}
}

func SetupHandlers() {
	http.HandleFunc("/mental-rotation/tasks", handleGetTasks)
	http.HandleFunc("/mental-rotation/submit", handleSubmitResult)
	http.HandleFunc("/mental-rotation/results", handleGetResults)

	// Create a sub-filesystem for the images directory
	imagesFS, err := fs.Sub(images, "images")
	if err != nil {
		panic(err)
	}
	http.Handle("/mental-rotation/images/", http.StripPrefix("/mental-rotation/images/", http.FileServer(http.FS(imagesFS))))

	http.Handle("/mental-rotation", http.HandlerFunc(serveMentalRotation))
}

func serveMentalRotation(w http.ResponseWriter, r *http.Request) {
	w.Write(frontendFile)
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func saveResults() error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(resultsFile, data, 0644)
}

func handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	results = append(results, result)
	err := saveResults()
	mu.Unlock()

	if err != nil {
		http.Error(w, "Failed to save results", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
