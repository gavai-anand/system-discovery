package dto

type PeersListResponse struct {
	Peers []string `json:"data"`
}

type CountResponse struct {
	Data struct {
		Value int `json:"value"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type OperationsResponse struct {
	Data    []IncrementRequest `json:"data"`
	Message string             `json:"message"`
	Success bool               `json:"success"`
}
