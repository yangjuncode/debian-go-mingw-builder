package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// fileInfo holds path and modification time for sorting
type fileInfo struct {
	path    string
	modTime time.Time
}

func main() {
	// --- Determine Target Directory ---
	targetDir := "." // Default to current directory
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}

	// Get absolute path for clearer messages and consistency
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		log.Fatalf("Error getting absolute path for '%s': %v", targetDir, err)
	}

	// --- Validate Directory ---
	dirInfo, err := os.Stat(absTargetDir)
	if os.IsNotExist(err) {
		log.Fatalf("Error: Directory '%s' does not exist.", absTargetDir)
	}
	if err != nil {
		log.Fatalf("Error accessing directory '%s': %v", absTargetDir, err)
	}
	if !dirInfo.IsDir() {
		log.Fatalf("Error: '%s' is not a directory.", absTargetDir)
	}

	fmt.Printf("Processing .diff files in directory '%s'...\n", absTargetDir)

	// --- Find and Collect .diff Files ---
	var diffFiles []fileInfo

	// Read directory entries
	entries, err := os.ReadDir(absTargetDir)
	if err != nil {
		log.Fatalf("Error reading directory '%s': %v", absTargetDir, err)
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		//if file name has begin with two digits and a dash, skip it
		fileName := entry.Name()
		if len(fileName) >= 3 && fileName[0] >= '0' && fileName[0] <= '9' && fileName[1] >= '0' && fileName[1] <= '9' && fileName[2] == '-' {
			fmt.Printf("Skipping file '%s' because it starts with two digits and a dash.\n", fileName)
			continue
		}

		// Check if it's a .diff or .patch file (case-insensitive check is often good)
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".diff") || strings.HasSuffix(strings.ToLower(entry.Name()), ".patch") {
			filePath := filepath.Join(absTargetDir, entry.Name())
			info, err := entry.Info() // Get FileInfo (includes mod time)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not get info for file '%s': %v. Skipping.\n", filePath, err)
				continue // Skip this file if we can't get its info
			}
			diffFiles = append(diffFiles, fileInfo{path: filePath, modTime: info.ModTime()})
		}
	}

	// --- Sort Files by Modification Time ---
	sort.Slice(diffFiles, func(i, j int) bool {
		// Sort oldest first
		return diffFiles[i].modTime.Before(diffFiles[j].modTime)
	})

	// --- Rename Files Sequentially ---
	renamedCount := 0
	for i, file := range diffFiles {
		dirName := filepath.Dir(file.path)
		baseName := filepath.Base(file.path)

		// Format prefix (e.g., 00-, 01-, ...)
		prefix := fmt.Sprintf("%02d-", i)
		newBaseName := prefix + baseName
		newPath := filepath.Join(dirName, newBaseName)

		// Safety check: does the new name already exist?
		if _, err := os.Stat(newPath); err == nil {
			// File exists
			fmt.Fprintf(os.Stderr, "Warning: Target file '%s' already exists. Skipping rename for '%s'.\n", newPath, baseName)
			continue
		} else if !os.IsNotExist(err) {
			// Different error occurred trying to stat the new path
			fmt.Fprintf(os.Stderr, "Warning: Could not check existence of target file '%s': %v. Skipping rename for '%s'.\n", newPath, err, baseName)
			continue
		}

		// Perform the rename
		fmt.Printf("Renaming: '%s' -> '%s'\n", baseName, newBaseName)
		err := os.Rename(file.path, newPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error renaming file '%s' to '%s': %v\n", file.path, newPath, err)
			// Decide if you want to stop or continue on error. Let's continue.
		} else {
			renamedCount++
		}
	}

	// --- Completion Message ---
	if len(diffFiles) == 0 { // Check original count, not renamedCount
		fmt.Printf("No .diff or .patch files need processing found in '%s'.\n", absTargetDir)
	} else {
		fmt.Printf("Processing complete. Attempted to rename %d files, successfully renamed %d.\n", len(diffFiles), renamedCount)
	}
}
