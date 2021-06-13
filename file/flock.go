package file

import (
	"os"
	"syscall"
)

//	apply or remove an advisory lock on an open file

// Flock
func Flock(file *os.File) error {
	// Note that this is non blocking.
	// If the file lock is currently held by someone, an error will be returned directly.
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func Funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
