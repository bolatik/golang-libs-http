package http

type Params map[string]interface{}

type Pagination struct {
	Page         int
	ItemsPerPage int
}

type ListParams struct {
	Query      Params
	Seq        *Sequence
	Pagination Pagination
	Sort       map[string]bool
	Fields     []string
}

type ListResponse struct {
	Ipp   int           `json:"ipp"`
	Page  int           `json:"p"`
	Total int64         `json:"total"`
	Items []interface{} `json:"items"`
}

type SequenceType int

const (
	SequenceNone SequenceType = iota
	SequenceFirst
	SequenceLast
)

var SequenceTypes = map[SequenceType]string{
	SequenceNone:  "none",
	SequenceFirst: "first",
	SequenceLast:  "last",
}

func (a SequenceType) String() string {
	switch a {
	default:
		return SequenceTypes[SequenceNone]
	case SequenceFirst:
		return SequenceTypes[SequenceFirst]
	case SequenceLast:
		return SequenceTypes[SequenceLast]
	}
	return SequenceTypes[SequenceNone]
}

type Sequence struct {
	Seq  SequenceType
	Size int
}
