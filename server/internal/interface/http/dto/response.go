// Package dto 定义 HTTP 请求和响应的数据传输对象。
package dto

// ApiResponse 统一 API 响应格式
type ApiResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) ApiResponse {
	return ApiResponse{Code: 0, Data: data, Message: "ok"}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) ApiResponse {
	return ApiResponse{Code: code, Data: nil, Message: message}
}

// PaginatedData 分页数据
type PaginatedData struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// PaginatedResponse 分页成功响应
func PaginatedResponse(items interface{}, total, limit, offset int) ApiResponse {
	return SuccessResponse(PaginatedData{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
