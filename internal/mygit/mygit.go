package mygit

const (
	DefaultGitDirectory = ".git"
	ObjectsDirectory    = "objects"
	RefsDirectory       = "refs"
)

type (
	MyGit struct {
		gitDirectory string
	}
	Opt func(m *MyGit)
)

func WithGitDirectory(name string) Opt {
	return func(m *MyGit) {
		m.gitDirectory = name
	}
}

func NewMyGit(opts ...Opt) *MyGit {
	m := &MyGit{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}
