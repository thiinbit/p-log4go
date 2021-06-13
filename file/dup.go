// +build !arm64

package file

import "golang.org/x/sys/unix"

//	dup, dup2, dup3 - duplicate a file descriptor

// SyscallDup on Unix/Linux
func SyscallDup(oldfd int, newfd int) (err error) {
	return unix.Dup2(oldfd, newfd)
}
