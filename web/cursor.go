package web

type CursorOptions struct {
	Cursor string
	Limit  int
}

type Cursor struct {
	NextCursor string `json:"nextCursor"`
}
