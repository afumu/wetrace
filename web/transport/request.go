package transport

// PaginationQuery 定义了列表请求的通用分页参数。
type PaginationQuery struct {
	Limit  int `form:"limit,default=20"`
	Offset int `form:"offset,default=0"`
}

// KeywordQuery 定义了一个通用的搜索参数。
type KeywordQuery struct {
	Keyword string `form:"keyword"`
}
