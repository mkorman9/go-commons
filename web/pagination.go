package web

type PaginationOptions struct {
	PageNumber int
	PageSize   int
}

func (opts PaginationOptions) Offset() int {
	return opts.PageNumber * opts.PageSize
}

func (opts PaginationOptions) Limit() int {
	return opts.PageSize
}

func (opts PaginationOptions) RealPageSize(recordsCount int64) int {
	pageSize := opts.PageSize
	if int(recordsCount) < pageSize {
		pageSize = int(recordsCount)
	}

	return pageSize
}

func (opts PaginationOptions) NumberOfPages(recordsCount int64) int {
	if recordsCount == 0 {
		return 0
	}

	splitToPagesNum := opts.PageSize
	if splitToPagesNum == 0 {
		splitToPagesNum = int(recordsCount)
	}

	pagesCount := int(recordsCount / int64(splitToPagesNum))
	if recordsCount%int64(splitToPagesNum) != 0 {
		pagesCount += 1
	}

	return pagesCount
}

type Page struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

type SortingOptions struct {
	Field   int
	Reverse bool
}
