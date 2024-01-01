# MyGit

## About

Playing around writing a toy git

## Progress

Init a git repository:
```
$ go run main.go --path /tmp/test-dir --git-directory .git init
$ git -C /tmp/test-dir status
On branch main

No commits yet

nothing to commit (create/copy files and use "git add" to track)
```


Create a file and commit to `main`:
```
$ echo "hello" > /tmp/test-dir/hello
$ go run main.go --path /tmp/test-dir --git-directory .git commit
$ git -C /tmp/test-dir show
```

Create another file and commit to `main` 
```
$ echo "world" > /tmp/test-dir/world
$ go run main.go --path /tmp/test-dir --git-directory .git commit
$ git -C /tmp/test-dir show
```

Add nested directory with files and commit to `main`:

```
$ mkdir -p /tmp/test-dir/nested
$ echo "file1" > /tmp/test-dir/nested/file1
$ echo "file2" > /tmp/test-dir/nested/file2
$ go run main.go --path /tmp/test-dir --git-directory .git commit
$ git -C /tmp/test-dir cat-file -p <the tree hash>
100644 blob ce013625030ba8dba906f756967f9e9ca394464a    hello
040000 tree 9f28c263e4723d7458338b0403f110a1e435de11    nested
100644 blob cc628ccd10742baea8241c5924df992b5c019f71    world
$ git -C /tmp/test-dir cat-file -p <the nested tree hash>
100644 blob e2129701f1a4d54dc44f03c93bca0a2aec7c5449    file1
100644 blob 6c493ff740f9380390d5c9ddef4af18697ac9375    file2
$ git -C /tmp/test-dir show
```


## Notes

`git cat-file` should be the first thing any explanation of git introduces for interrogating Git internals. It allows 
you to inspect object files trivially and with nice formatting.
