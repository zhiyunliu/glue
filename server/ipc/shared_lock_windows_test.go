//go:build windows

package ipc

func cleanupShared(string) error {
	return nil
}
