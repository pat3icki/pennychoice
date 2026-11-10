package cache

type Options interface {
	Validate() error
	Get(string) any
}

type Cache interface {
	Store(key, data []byte, opts Options)
	Get(key, data []byte)
	Delete(key string)
}
