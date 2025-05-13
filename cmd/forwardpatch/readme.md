# forwardpatch

forwardpatch is a tool to apply patch files to the specified branch from another branch.
they have the same common ancestor, but divergence from the common ancestor is different.

## Usage

forwardpatch <git-repo-path>  <from-branch> <to-branch> <base-branch>

- git-repo-path is the path to the git repository
- from-branch is the branch to extract the patch from
- to-branch is the branch to apply the patch to
- base-branch is the base branch to find the common ancestor

## Example

forwardpatch <go-git-repo-path> 1.23.8 1.23.9 release-go-1.23

## Tech

use git merge-base to determine the common ancestor of the from-branch and the base-branch, then use git rev-list --reverse to get all commits from the common ancestor to the from-branch, then cherry pick the commits to the to-branch.
