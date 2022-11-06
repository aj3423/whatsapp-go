package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

const defaultFailedCode = -3423

func RunCommand(dir, name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	if len(dir) > 0 {
		cmd.Dir = dir
	}

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			exitCode = defaultFailedCode
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return
}

func convert(exif, jfif string) bool {
	stdout, stderr, exit_code := RunCommand(".", "python3", "convert.py", exif, jfif)
	_ = stdout
	_ = stderr
	//fmt.Println(stdout, stderr)
	return exit_code == 0
}

func main() {
	exif := "/e/jpg.jpg"
	jfif := "/e/jfif.jfif"

	err := convert(exif, jfif)
	fmt.Println(err)
}
