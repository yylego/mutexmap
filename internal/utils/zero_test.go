// Package utils_test: Test suite for generic utility functions
// Tests zero value generation across various Go types
// Validates type-safe initialization patterns with Go generics
//
// utils_test: 泛型工具函数的测试套件
// 测试各种 Go 类型的零值生成
// 验证使用 Go 泛型的类型安全初始化模式
package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/mutexmap/internal/utils"
)

// TestZero validates Zero function returning correct zero values across multiple types
// Tests integer, string, floating-point, unsigned, and boolean zero value generation
// Ensures generic type parameter mechanism produces correct default values
//
// TestZero 验证 Zero 函数跨多个类型返回正确的零值
// 测试整数、字符串、浮点数、无符号和布尔零值生成
// 确保泛型类型参数机制产生正确的默认值
func TestZero(t *testing.T) {
	res := utils.Zero[int]()
	t.Log(res)
	require.Zero(t, res)

	require.Zero(t, utils.Zero[string]())
	require.Zero(t, utils.Zero[float64]())
	require.Zero(t, utils.Zero[uint32]())
	require.Zero(t, utils.Zero[bool]())
	require.Zero(t, utils.Zero[int64]())
}
