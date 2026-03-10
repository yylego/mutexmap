// Package utils provides generic utilities and helper functions
// Supports common operations used across mutexmap implementations
//
// utils 包提供泛型工具和辅助函数
// 支持 mutexmap 实现中使用的通用操作
package utils

// Zero returns the zero value for any type T
// Uses Go's default value initialization through named return parameter
// Useful when a zero value is needed as a placeholder or default
//
// Zero 返回任意类型 T 的零值
// 通过命名返回参数使用 Go 的默认值初始化
// 在需要零值作为占位符或默认值时很有用
func Zero[T any]() (x T) {
	return x
}
