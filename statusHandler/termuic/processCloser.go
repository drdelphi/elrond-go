package termuic

import (
	"os"
	"runtime"
	"syscall"
)

// StopApplication is used to stop aplication when Ctrl+C is pressed
func StopApplication() {
	if p, err := os.FindProcess(os.Getpid()); err != nil {
		return
	} else {
		if runtime.GOOS == "windows" {
			_ = p.Kill()
		} else {
			_ = p.Signal(syscall.SIGINT)
		}
	}
}
