package dto

type RegisterRequest struct {
	NodeID string `json:"node_id" validate:"required"`
}

type SyncPeersRequest struct {
	Peers []string `json:"peers" validate:"required"`
}

type IncrementRequest struct {
	ID     int64  `json:"id"`
	Source string `json:"source"`
	Value  int    `json:"value"`
}
