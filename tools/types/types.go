// Package types implements some commonly used db serializable types like datetime, json, etc.
package types

// This is a wild function, kill me pls
func Pointer[T any](val T) *T {
	return &val
}
