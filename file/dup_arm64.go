// +build arm64

package file

import (
	"syscall"
)

// SyscallDup on Arm64
func SyscallDup(oldfd int, newfd int) (err error) {
	// linux_arm64 platform doesn't have syscall.Dup2
	// so use the nearly identical syscall.Dup3 instead.
	return syscall.Dup3(oldfd, newfd, 0)
}
