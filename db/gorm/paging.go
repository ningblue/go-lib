package gorm

// NormalizePage 分页参数归一：pageNum<1 → 1；pageSize<1 → 默认 20；pageSize>200 → 200。
func NormalizePage(pageNum, pageSize int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageNum, pageSize
}
