package model

type DiffOp int

const (
	DiffEqual DiffOp = iota
	DiffInsert
	DiffDelete
	DiffComment // HTML comment fragment inside the diff (gray-styled in UI)
)

type DiffSegment struct {
	Op   DiffOp
	Text string
}

type Commit struct {
	Hash       string
	ShortHash  string
	Message    string
	FirstLine  string
	HasMore    bool
	Timestamp  int64
	IsUnstaged bool // 顶部"未提交"特殊节点
}

type LoadResult struct {
	Path           string
	Content        string
	Title          string
	Summary        string
	HasFrontmatter bool
	CharCount      int
}

type Result struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}