package router

import "testing"

func TestGroup_GetReallyPath(t *testing.T) {
	// 创建 Group 结构体的实例
	group := &Group{
		rootPath: "/root",
		Path:     "example/path",
	}

	// 测试 GetReallyPath() 方法
	expected := "/root/example/path"
	result := group.GetReallyPath()
	if result != expected {
		t.Errorf("GetReallyPath() = %s; want %s", result, expected)
	}

	// 测试 rootPath 为空的情况
	group.rootPath = ""
	expected = "/example/path"
	result = group.GetReallyPath()
	if result != expected {
		t.Errorf("GetReallyPath() = %s; want %s", result, expected)
	}
}
