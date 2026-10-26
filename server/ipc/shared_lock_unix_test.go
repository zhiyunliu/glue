//go:build linux || darwin

package ipc

func cleanupShared(name string) error {
	return unlinkSharedObject("/" + name)
}
