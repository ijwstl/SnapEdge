package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Constants for border location
const (
	BOTTOM = "bottom"
	TOP    = "top"
	LEFT   = "left"
	RIGHT  = "right"
	ALL    = "all"
)

// Config structs with clear field documentation
type Config struct {
	ImagePath  string `json:"imagePath"`  // Path to source images
	OutputPath string `json:"outputPath"` // Path to save processed images
	Quality    int    `json:"quality"`    // JPEG quality (1-100)
	TaskMax    int    `json:"taskMax"`    // Maximum concurrent tasks
	Border     Border `json:"border"`     // Border configuration
	Logo       Logo   `json:"logo"`       // Logo configuration
	Text       []Text `json:"text"`       // Text overlay configuration
}

type Border struct {
	BorderWidth        float64 `json:"borderWidth"`        // Width of left/right borders as percentage of image width
	BorderBottomHeight float64 `json:"borderBottomHeight"` // Height of bottom border as percentage of image height
	BorderTopHeight    float64 `json:"borderTopHeight"`    // Height of top border as percentage of image height
	BorderLocation     string  `json:"borderLocation"`     // Border placement: "top", "bottom", "left", "right", "all"
	BorderColor        string  `json:"borderColor"`        // Border color
}

type Logo struct {
	On       bool   `json:"on"`       // Enable logo overlay
	FilePath string `json:"filePath"` // Path to logo image file, "default" uses camera brand logo
	Resize   string `json:"resize"`   // Resize factor, "auto" or float value
}

type Text struct {
	On       bool   `json:"on"`       // Enable text overlay
	Text     string `json:"text"`     // Text content or special key: "device", "lens", "expose", "time"
	FontPath string `json:"fontPath"` // Path to font file
	FontSize string `json:"fontSize"` // Font size, "auto" or numeric value
	Bold     int    `json:"bold"`     // Bold level (0 = normal)
	Location string `json:"location"` // Text placement: "UpperLeft", "LowerLeft", etc.
}

// Task represents a single image processing job
type Task struct {
	FilePath    string // Source image path
	OutFilePath string // Destination image path
}

// Main entry point
func main() {
	// Read configuration
	config, err := readConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	fmt.Println("Processing with configuration:", config)

	// Create a worker pool for concurrent image processing
	var wg sync.WaitGroup
	taskChannel := make(chan Task, config.TaskMax)

	// Start worker pool
	startWorkers(&wg, taskChannel, config)

	// Scan directories and enqueue tasks
	if err := scanAndEnqueueTasks(config.ImagePath, config.OutputPath, &wg, &taskChannel); err != nil {
		fmt.Println("Error scanning files:", err)
	}

	// Signal that no more tasks will be sent
	close(taskChannel)

	// Wait for all tasks to complete
	wg.Wait()

	fmt.Println("All images processed and saved to:", config.OutputPath)
}

// startWorkers creates a pool of worker goroutines to process images
func startWorkers(wg *sync.WaitGroup, taskChannel chan Task, config Config) {
	go func() {
		for task := range taskChannel {
			go func(task Task) {
				defer wg.Done()
				err := AddWhiteBorderWithText(task.FilePath, task.OutFilePath, config)
				if err != nil {
					fmt.Printf("Error processing image %s: %v\n", task.FilePath, err)
				}
			}(task)
		}
	}()
}

// scanAndEnqueueTasks recursively scans directories for images and adds them to the task queue
func scanAndEnqueueTasks(imagePath, outputPath string, wg *sync.WaitGroup, taskChannel *chan Task) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputPath, err)
	}

	// Get file/directory info
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return fmt.Errorf("error accessing path %s: %w", imagePath, err)
	}

	// Handle directory
	if fileInfo.IsDir() {
		files, err := os.ReadDir(imagePath)
		if err != nil {
			return fmt.Errorf("error reading directory %s: %w", imagePath, err)
		}

		for _, file := range files {
			filePath := filepath.Join(imagePath, file.Name())

			// Recursively process subdirectories
			if file.IsDir() {
				subOutputPath := filepath.Join(outputPath, file.Name())
				if err := scanAndEnqueueTasks(filePath, subOutputPath, wg, taskChannel); err != nil {
					fmt.Printf("Warning: Error processing subdirectory %s: %v\n", filePath, err)
					continue
				}
			} else if isImageFile(file.Name()) {
				// Add image task to queue
				wg.Add(1)
				*taskChannel <- Task{
					FilePath:    filePath,
					OutFilePath: filepath.Join(outputPath, file.Name()),
				}
			}
		}
		return nil
	}

	// Handle single file
	if fileInfo.Mode().IsRegular() && isImageFile(fileInfo.Name()) {
		wg.Add(1)
		*taskChannel <- Task{
			FilePath:    imagePath,
			OutFilePath: filepath.Join(outputPath, fileInfo.Name()),
		}
		return nil
	}

	return fmt.Errorf("path %s is neither a supported image file nor a directory", imagePath)
}

// isImageFile checks if a filename has an image extension
func isImageFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// readConfig loads the application configuration from config.json
func readConfig() (Config, error) {
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config JSON: %w", err)
	}

	return config, nil
}
