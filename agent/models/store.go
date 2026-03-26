package models

type Store interface {
	BasePath() string
	ReadMeta(key string, v any) error
	ReadData(key string) ([]byte, error)
	Write(key string, data []byte, metadata any) error
	Delete(key string) error
	DeleteAll() error
}
