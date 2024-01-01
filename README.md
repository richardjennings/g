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
$ git -C /tmp/test-dir cat-file -p 5643ef0140306f73a16a1e903da8fcff9a5a9602
100644 blob 2c3ae82e5e5516b801382fc0efdb50e9a05c2430    hello
040000 tree 9f28c263e4723d7458338b0403f110a1e435de11    nested
100644 blob 2c3ae82e5e5516b801382fc0efdb50e9a05c2430    world
$ git -C /tmp/test-dir cat-file -p 9f28c263e4723d7458338b0403f110a1e435de11
100644 blob 2c3ae82e5e5516b801382fc0efdb50e9a05c2430    file1
100644 blob 2c3ae82e5e5516b801382fc0efdb50e9a05c2430    file2
```


## Notes

`git cat-file` should be the first thing any explanation of git introduces for interrogating Git internals. It allows 
you to inspect object files trivially and with nice formatting.
