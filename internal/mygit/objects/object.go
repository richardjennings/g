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
)

const (
	ObjectInvalid objectType = iota
	ObjectBlob
	ObjectTree
	ObjectCommit
)
