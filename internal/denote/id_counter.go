package denote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// IDCounter manages sequential IDs for tasks
type IDCounter struct {
	NextTaskID int `json:"next_task_id"`
	mu         sync.Mutex
	filePath   string
}

var (
	globalCounter     *IDCounter
	globalCounterOnce sync.Once
)

// GetIDCounter returns the singleton ID counter for the given tasks directory
func GetIDCounter(tasksDir string) (*IDCounter, error) {
	var err error
	globalCounterOnce.Do(func() {
		globalCounter, err = loadOrCreateCounter(tasksDir)
	})
	return globalCounter, err
}

// loadOrCreateCounter loads an existing counter or creates a new one
func loadOrCreateCounter(tasksDir string) (*IDCounter, error) {
	counterFile := filepath.Join(tasksDir, ".notes-cli-id-counter.json")
	
	// Try to load existing counter
	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Counter doesn't exist, scan for highest ID
			maxID := findMaxTaskID(tasksDir)
			
			counter := &IDCounter{
				NextTaskID: maxID + 1,
				filePath:   counterFile,
			}
			
			// Save the initial counter
			if err := counter.save(); err != nil {
				return nil, fmt.Errorf("failed to save initial counter: %w", err)
			}
			
			return counter, nil
		}
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}
	
	// Parse existing counter
	var counter IDCounter
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, fmt.Errorf("failed to parse counter file: %w", err)
	}
	
	counter.filePath = counterFile
	return &counter, nil
}

// findMaxTaskID scans the directory for the highest task ID
func findMaxTaskID(tasksDir string) int {
	maxID := 0
	
	pattern := filepath.Join(tasksDir, "*__task*.md")
	files, _ := filepath.Glob(pattern)
	
	for _, file := range files {
		task, err := ParseTaskFile(file)
		if err != nil {
			continue
		}
		if task.TaskID > maxID {
			maxID = task.TaskID
		}
	}
	
	return maxID
}

// NextID returns the next task ID and increments the counter
func (c *IDCounter) NextID() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.NextTaskID
	c.NextTaskID++
	
	if err := c.save(); err != nil {
		// Rollback on save failure
		c.NextTaskID--
		return 0, fmt.Errorf("failed to save counter: %w", err)
	}
	
	return id, nil
}

// save writes the counter to disk
func (c *IDCounter) save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}
	
	// Write to temp file first for atomicity
	tempFile := c.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename temp file to actual file (atomic on most systems)
	if err := os.Rename(tempFile, c.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename counter file: %w", err)
	}
	
	return nil
}

// ResetSingleton resets the singleton (useful for testing or config changes)
func ResetSingleton() {
	globalCounterOnce = sync.Once{}
	globalCounter = nil
}