# makepatch

makepatch is a tool to make patch files for git branch where it diverges from the specified branch.

## Usage

makepatch <git-repo-path> <current-branch> <base-branch> <output-directory>

- git-repo-path: path to the git repository
- current-branch: current branch, the need to make patch branch
- base-branch: base branch, the branch to compare with
- output-directory: output directory, the directory to store the patch files, every commit will be stored in a separate file, like 00-commit-msg.patch, 01-commit-msg.patch, etc.

## Example

makepatch <go-git-repo-path> 1.23.9 release-go-1.23 patch-dir

## tech

use git merge-base to determine the common ancestor of the current branch and the base branch, then use git rev-list --reverse to get all commits from the common ancestor to the current branch, then use git format-patch to generate patch files.