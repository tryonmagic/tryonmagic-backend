package service

func GetPaginationOffset(page int, pageSize int) int {
	return (page - 1) * pageSize
}

func CalculateTotalPages(total int64, pageSize int) int64 {
	return (total + int64(pageSize) - 1) / int64(pageSize)
}
