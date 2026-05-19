package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	HistoryMaxSize = 100
	HistoryFile    = ".ai-cmd-history.json"
)

// HistoryEntry represents a single command history entry
type HistoryEntry struct {
	ID        int       `json:"id"`
	Query     string    `json:"query"`
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
}

// HistoryManager manages command history
type HistoryManager struct {
	entries []HistoryEntry
	filePath string
	nextID   int
}

// NewHistoryManager creates a new history manager
func NewHistoryManager() (*HistoryManager, error) {
	hm := &HistoryManager{
		entries: make([]HistoryEntry, 0),
	}

	// Get home directory
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	hm.filePath = filepath.Join(usr.HomeDir, HistoryFile)

	// Load existing history
	if err := hm.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load history: %v", err)
	}

	return hm, nil
}

// load reads history from file
func (hm *HistoryManager) load() error {
	data, err := os.ReadFile(hm.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &hm.entries); err != nil {
		return fmt.Errorf("failed to parse history: %v", err)
	}

	// Find max ID
	hm.nextID = 0
	for _, entry := range hm.entries {
		if entry.ID >= hm.nextID {
			hm.nextID = entry.ID + 1
		}
	}

	return nil
}

// save writes history to file
func (hm *HistoryManager) save() error {
	// Ensure directory exists
	dir := filepath.Dir(hm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %v", err)
	}

	data, err := json.MarshalIndent(hm.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %v", err)
	}

	if err := os.WriteFile(hm.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history: %v", err)
	}

	return nil
}

// Add adds a new entry to history
func (hm *HistoryManager) Add(query, command string, success bool) {
	entry := HistoryEntry{
		ID:        hm.nextID,
		Query:     query,
		Command:   command,
		Timestamp: time.Now(),
		Success:   success,
	}

	hm.nextID++
	hm.entries = append(hm.entries, entry)

	// Keep only last N entries
	if len(hm.entries) > HistoryMaxSize {
		hm.entries = hm.entries[len(hm.entries)-HistoryMaxSize:]
	}

	// Save to file
	if err := hm.save(); err != nil {
		fmt.Fprintf(os.Stderr, "Debug: Failed to save history: %v\n", err)
	}
}

// GetAll returns all history entries
func (hm *HistoryManager) GetAll() []HistoryEntry {
	return hm.entries
}

// GetLast returns the last entry
func (hm *HistoryManager) GetLast() *HistoryEntry {
	if len(hm.entries) == 0 {
		return nil
	}
	return &hm.entries[len(hm.entries)-1]
}

// GetByID returns entry by ID
func (hm *HistoryManager) GetByID(id int) *HistoryEntry {
	for i := len(hm.entries) - 1; i >= 0; i-- {
		if hm.entries[i].ID == id {
			return &hm.entries[i]
		}
	}
	return nil
}

// Search searches for entries containing keyword
func (hm *HistoryManager) Search(keyword string) []HistoryEntry {
	var results []HistoryEntry
	keyword = strings.ToLower(keyword)

	for i := len(hm.entries) - 1; i >= 0; i-- {
		entry := hm.entries[i]
		if strings.Contains(strings.ToLower(entry.Query), keyword) ||
			strings.Contains(strings.ToLower(entry.Command), keyword) {
			results = append(results, entry)
		}
	}

	return results
}

// PrintHistory prints formatted history
func (hm *HistoryManager) PrintHistory() {
	if len(hm.entries) == 0 {
		fmt.Println("No command history found.")
		return
	}

	fmt.Println("\n📜 Command History (most recent first):")
	fmt.Println(strings.Repeat("─", 80))

	// Sort by timestamp descending
	sorted := make([]HistoryEntry, len(hm.entries))
	copy(sorted, hm.entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	for i := len(sorted) - 1; i >= 0; i-- {
		entry := sorted[i]
		status := "✓"
		if !entry.Success {
			status = "✗"
		}

		timeStr := entry.Timestamp.Format("2006-01-02 15:04")
		fmt.Printf("[%3d] %s | %s\n", entry.ID, timeStr, entry.Query)
		fmt.Printf("     └─> %s %s\n", status, entry.Command)
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("Total: %d commands\n", len(hm.entries))
}

// ResolveSpecialCommand resolves special history commands like !!, !123, !keyword
func (hm *HistoryManager) ResolveSpecialCommand(query string) (string, bool) {
	query = strings.TrimSpace(query)

	// Handle !! - repeat last command
	if query == "!!" {
		last := hm.GetLast()
		if last == nil {
			fmt.Println("Error: No previous command in history.")
			return "", false
		}
		fmt.Printf("Repeating last command: %s\n", last.Query)
		return last.Query, true
	}

	// Handle !<number> - repeat command by ID
	if len(query) > 1 && query[0] == '!' {
		rest := query[1:]
		
		// Try to parse as number
		if id, err := parseInt(rest); err == nil {
			entry := hm.GetByID(id)
			if entry == nil {
				fmt.Printf("Error: Command with ID %d not found.\n", id)
				return "", false
			}
			fmt.Printf("Repeating command #%d: %s\n", id, entry.Query)
			return entry.Query, true
		}

		// Search by keyword
		if rest != "" {
			results := hm.Search(rest)
			if len(results) == 0 {
				fmt.Printf("Error: No commands found matching '%s'.\n", rest)
				return "", false
			}
			// Return most recent match
			fmt.Printf("Found match: %s\n", results[0].Query)
			return results[0].Query, true
		}
	}

	return "", false
}

// parseInt parses a string to integer
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit")
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
