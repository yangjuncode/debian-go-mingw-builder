package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	// Check if the correct number of arguments is provided
	if len(os.Args) < 4 || len(os.Args) > 5 {
		fmt.Println("Usage: makepatch <git-repo-path> <current-branch> <base-branch> [output-directory]")
		fmt.Println("  output-directory is optional, default is 'patch'")
		os.Exit(1)
	}

	// Parse command-line arguments
	gitRepoPath := os.Args[1]
	currentBranch := os.Args[2]
	baseBranch := os.Args[3]
	outputDir := "patch" // Default value

	// Use provided output directory if specified
	if len(os.Args) == 5 {
		outputDir = os.Args[4]
	}

	// Convert output directory to absolute path if it's not already
	if !filepath.IsAbs(outputDir) {
		absOutputDir, err := filepath.Abs(outputDir)
		if err != nil {
			fmt.Printf("Error getting absolute path for output directory: %v\n", err)
			os.Exit(1)
		}
		outputDir = absOutputDir
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Change to the git repository directory
	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gitRepoPath); err != nil {
		fmt.Printf("Error changing to git repository directory: %v\n", err)
		os.Exit(1)
	}

	// Find the common ancestor of the current branch and the base branch
	cmd := exec.Command("git", "merge-base", currentBranch, baseBranch)
	commonAncestor, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error finding common ancestor: %v\n", err)
		os.Exit(1)
	}
	ancestorSHA := strings.TrimSpace(string(commonAncestor))

	// Get all commits from the common ancestor to the current branch
	cmd = exec.Command("git", "rev-list", "--reverse", ancestorSHA+"..."+currentBranch)
	commits, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting commit list: %v\n", err)
		os.Exit(1)
	}

	commitList := strings.Split(strings.TrimSpace(string(commits)), "\n")
	if len(commitList) == 1 && commitList[0] == "" {
		fmt.Println("No commits found between the branches.")
		os.Exit(0)
	}

	// Generate patch files for each commit
	for i, commit := range commitList {
		// Get commit message for the filename
		cmd = exec.Command("git", "log", "-1", "--pretty=%s", commit)
		messageBytes, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error getting commit message: %v\n", err)
			os.Exit(1)
		}
		message := strings.TrimSpace(string(messageBytes))

		// Sanitize the commit message for use in a filename
		sanitizedMessage := sanitizeFilename(message)

		// Create the patch file
		patchFileName := fmt.Sprintf("%02d-%s.patch", i, sanitizedMessage)
		//replace _. to .
		for strings.Contains(patchFileName, "_.") {
			patchFileName = strings.ReplaceAll(patchFileName, "_.", ".")
		}

		patchFilePath := filepath.Join(outputDir, patchFileName)

		// Generate the patch
		cmd = exec.Command("git", "format-patch", "-1", "--stdout", commit)
		patchContent, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error generating patch for commit %s: %v\n", commit, err)
			os.Exit(1)
		}

		// Write the patch to file
		if err := os.WriteFile(patchFilePath, patchContent, 0644); err != nil {
			fmt.Printf("Error writing patch file %s: %v\n", patchFilePath, err)
			os.Exit(1)
		}

		fmt.Printf("Created patch file: %s\n", patchFilePath)
	}

	fmt.Printf("Successfully created %d patch files in %s\n", len(commitList), outputDir)
}

// sanitizeFilename removes or replaces characters that are not allowed in filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"<", ">", "(", ")", ":", "\"", "/", "\\", "|", "?", "*", ".", " "}
	result := filename

	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// replace all unicode punctuations to _
	result = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return '_'
		}
		return r
	}, result)

	// replace __ to _
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// replace _. to .
	for strings.Contains(result, "_.") {
		result = strings.ReplaceAll(result, "_.", ".")
	}

	// Truncate long filenames
	if len(result) > 72 {
		result = result[:72]
	}

	return result
}
