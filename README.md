# G

Learning Git internals by writing a Git in Go.

## CLI

```
go run ./cmd/g --help
```

## Library

```go
// new repo
repo, _ := g.Init(g.WithPath("/tmp/myrepo"))

// existing repo
repo, _ := g.Open(g.WithPath("/tmp/myrepo"))

// status, commit
status, _ := repo.CurrentStatus()
repo.CreateCommit(&g.Commit{...})

// branches
repo.CreateBranch("feature")
repo.SwitchBranch("feature")
branches, _ := repo.ListBranches()
```
