# MyGit

## About

Playing around writing a toy git

## Progress

### First step

Init a git repository. Running `git status` in `test-dir` should work.

`./mygit init --path test-dir --git-directory .git`

Create a file `echo "hello" > test-dir/hello`

Adds a commit for `main`:
`./mygit commit --path test-dir --git-directory .git`

Git show should print out the commit.

Create another file `echo "world" > test-dir/world`

Add another commit for `main`:
`./mygit commit --path test-dir --git-directory .git`

`git log` should now show both commits.