package DTO

type DefaultResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type Response struct {
	Status string `json:"status"`
}
