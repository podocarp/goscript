package kind

type Kind int

const (
	// literals
	STRING Kind = iota
	FLOAT

	FUNC
	ARRAY
)

var kinds = []string{
	STRING: "STRING",
	FLOAT:  "FLOAT",
	FUNC:   "FUNC",
	ARRAY:  "ARRAY",
}

func (kind Kind) String() string {
	if 0 <= kind && kind < Kind(len(kinds)) {
		return kinds[kind]
	}
	return "unknown"
}
