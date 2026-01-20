package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Check if the correct number of arguments is provided
	if len(os.Args) != 5 {
		fmt.Println("Usage: forwardpatch <git-repo-path> <from-branch> <to-branch> <base-branch>")
		fmt.Println("  - git-repo-path is the path to the git repository")
		fmt.Println("  - from-branch is the branch to extract the patch from")
		fmt.Println("  - to-branch is the branch to apply the patch to")
		fmt.Println("  - base-branch is the base branch to find the common ancestor")
		os.Exit(1)
	}

	// Parse command-line arguments
	gitRepoPath := os.Args[1]
	fromBranch := os.Args[2]
	toBranch := os.Args[3]
	baseBranch := os.Args[4]

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

	// Verify that the branches exist
	if !branchExists(fromBranch) {
		fmt.Printf("Error: from-branch '%s' does not exist\n", fromBranch)
		os.Exit(1)
	}

	if !branchExists(baseBranch) {
		fmt.Printf("Error: base-branch '%s' does not exist\n", baseBranch)
		os.Exit(1)
	}

	if !branchExists(toBranch) {
		fmt.Printf("to-branch '%s' does not exist.\n", toBranch)
		fmt.Printf("Do you want to create it based on '%s'? [Y/n]: ", baseBranch)

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response == "" || response == "y" || response == "yes" {
			// Create the branch
			cmd := exec.Command("git", "checkout", "-b", toBranch, baseBranch)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error creating branch '%s': %v\n%s\n", toBranch, err, string(output))
				os.Exit(1)
			}
			fmt.Printf("Created branch '%s' based on '%s'\n", toBranch, baseBranch)
		} else {
			fmt.Println("Aborting.")
			os.Exit(1)
		}
	}

	// Find the common ancestor of the from-branch and the base-branch
	cmd := exec.Command("git", "merge-base", fromBranch, baseBranch)
	commonAncestor, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error finding common ancestor: %v\n", err)
		os.Exit(1)
	}
	ancestorSHA := strings.TrimSpace(string(commonAncestor))

	// Get all commits from the common ancestor to the from-branch
	cmd = exec.Command("git", "rev-list", "--reverse", ancestorSHA+"..."+fromBranch)
	commits, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting commit list: %v\n", err)
		os.Exit(1)
	}

	commitList := strings.Split(strings.TrimSpace(string(commits)), "\n")
	if len(commitList) == 1 && commitList[0] == "" {
		fmt.Println("No commits found between the common ancestor and from-branch.")
		os.Exit(0)
	}

	// Checkout to-branch
	cmd = exec.Command("git", "checkout", toBranch)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error checking out to-branch '%s': %v\n", toBranch, err)
		os.Exit(1)
	}

	// Cherry-pick each commit to the to-branch
	successCount := 0
	conflictCount := 0
	skipCount := 0

	for i, commit := range commitList {
		// Get commit message for logging
		cmd = exec.Command("git", "log", "-1", "--pretty=%s", commit)
		messageBytes, err := cmd.Output()
		commitMsg := "<unknown>"
		if err == nil {
			commitMsg = strings.TrimSpace(string(messageBytes))
		}

		fmt.Printf("[%d/%d] Cherry-picking commit %s: %s\n", i+1, len(commitList), commit[:8], commitMsg)

		// We'll skip the check for existing commits and try to cherry-pick directly
		// The cherry-pick command will fail if the changes are already applied

		// Cherry-pick the commit
		cmd = exec.Command("git", "cherry-pick", commit)
		cherryPickOutput, err := cmd.CombinedOutput()
		outputStr := string(cherryPickOutput)

		if err != nil {
			// Check if the patch is already applied
			if strings.Contains(outputStr, "patch does not apply") ||
				strings.Contains(outputStr, "patch already applied") ||
				strings.Contains(outputStr, "nothing to commit") ||
				strings.Contains(outputStr, "nothing new to commit") {
				fmt.Printf("  Skipping: Changes already applied in '%s'\n", toBranch)
				// Abort the cherry-pick
				abortCmd := exec.Command("git", "cherry-pick", "--abort")
				abortCmd.Run() // Ignore errors here
				skipCount++
				// Check if there's a conflict
			} else if strings.Contains(outputStr, "conflict") || strings.Contains(outputStr, "CONFLICT") {
				fmt.Printf("  Conflict detected. Aborting cherry-pick.\n")
				// Abort the cherry-pick
				abortCmd := exec.Command("git", "cherry-pick", "--abort")
				abortCmd.Run() // Ignore errors here
				conflictCount++
			} else {
				fmt.Printf("  Error cherry-picking commit: %v\n%s\n", err, outputStr)
				// Abort the cherry-pick
				abortCmd := exec.Command("git", "cherry-pick", "--abort")
				abortCmd.Run() // Ignore errors here
				conflictCount++
			}
		} else {
			fmt.Printf("  Successfully cherry-picked commit\n")
			successCount++
		}
	}

	// Summary
	fmt.Println("\nForward patch summary:")
	fmt.Printf("  Total commits: %d\n", len(commitList))
	fmt.Printf("  Successfully applied: %d\n", successCount)
	fmt.Printf("  Skipped (already exists): %d\n", skipCount)
	fmt.Printf("  Failed (conflicts): %d\n", conflictCount)
}

// branchExists checks if a branch exists in the repository
func branchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	err := cmd.Run()
	return err == nil
}
