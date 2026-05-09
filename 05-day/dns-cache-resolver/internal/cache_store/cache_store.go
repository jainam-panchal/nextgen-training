// Package cachestore
package cachestore

type CacheStore[T any] interface {
	Keys() []string
	Get(string) (T, bool)
	Set(string, T)
	Delete(string) bool
	Len() int
}
