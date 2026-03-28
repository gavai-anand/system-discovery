package interfaces

type ICacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}
