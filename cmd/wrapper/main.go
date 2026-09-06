// yted-wrapper is a tiny static wrapper that strips poisoned LD_LIBRARY_PATH
// before execing the real YTed binary. AppImages and snap base snaps export
// a stale LD_LIBRARY_PATH (e.g. /snap/core20/... with glibc 2.31 vs host
// 2.43) that poisons every child process. A shell wrapper fails because
// /bin/sh itself is poisoned; this static Go binary has no dynamic deps and
// can safely unset the env before exec.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	os.Unsetenv("LD_LIBRARY_PATH")
	os.Unsetenv("LD_PRELOAD")
	os.Unsetenv("SNAP_LIBRARY_PATH")
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir := filepath.Dir(exe)
	bin := filepath.Join(dir, "yted.bin")
	args := append([]string{bin}, os.Args[1:]...)
	env := os.Environ()
	if err := syscall.Exec(bin, args, env); err != nil {
		cmd := exec.Command(bin, os.Args[1:]...)
		cmd.Env = env
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err2 := cmd.Run(); err2 != nil {
			if exitErr, ok := err2.(*exec.ExitError); ok {
				if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(ws.ExitStatus())
				}
			}
			os.Exit(1)
		}
	}
}
