package objects

type (
	Object struct {
		Path    string
		Typ     objectType
		Sha     []byte
		Objects []*Object
		//mode    string
	}
	objectType int
	ObjectFile struct {
		Path string
		Sha  []byte
	}
)

const (
	ObjectInvalid objectType = iota
	ObjectBlob
	ObjectTree
	ObjectCommit
)
