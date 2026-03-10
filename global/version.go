package global

import (
	"errors"
	"fmt"
	"runtime/debug"
)

var (
	debugInfo *debug.BuildInfo
	ok        bool
)

// GetPackageVersion 获取指定包的版本信息
func GetPackageVersion(pkgPath string) (string, error) {
	if debugInfo == nil {
		debugInfo, ok = debug.ReadBuildInfo()
	}
	if !ok {
		return "", errors.New("无法读取构建信息")
	}

	// 遍历所有依赖
	for _, dep := range debugInfo.Deps {
		if dep.Path == pkgPath {
			// 返回版本号
			return dep.Version, nil
		}

		// 如果依赖被替换，检查替换的模块
		if dep.Replace != nil && dep.Replace.Path == pkgPath {
			return dep.Replace.Version, nil
		}
	}

	return "", fmt.Errorf("未找到包: %s", pkgPath)
}

// GetAllDependencies 获取所有依赖的版本
func GetAllDependencies() map[string]string {
	deps := make(map[string]string)
	if debugInfo == nil {
		debugInfo, ok = debug.ReadBuildInfo()
	}
	if !ok {
		return deps
	}

	for _, dep := range debugInfo.Deps {
		version := dep.Version
		if dep.Replace != nil {
			version = dep.Replace.Version
		}
		deps[dep.Path] = version
	}

	return deps
}
