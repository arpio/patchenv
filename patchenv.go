package patchenv

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// patchCommandVar is the environment variable used by Patch that, when set,
// contains the command that is run to compute the new environment variables
// to set in the running process.
const patchCommandVar = "PATCH_ENV_COMMAND"

// shellVar is the name of the environment variable that indicates the user's
// configured shell.
const shellVar = "SHELL"

// Patch checks if the PATCH_ENV_COMMAND environment variable is set, and if it
// is, runs it with the current shell (indicated by the SHELL environment
// variable), parses output lines as "var=value", and sets each "var" to
// "value" using os.Setenv() in the current process.  An error is returned if
// the command could not be run or exits with an error status.
//
// If the command returns an error status, the command's stdout and stderr
// are written to os.Stdout and os.Stderr respectively to help the user
// diagnose the problem.  Otherwise, the command's stderr is discarded and
// the command's stdout is parsed for the environment variables to set in
// the running process.
//
// If PATCH_ENV_COMMAND is not set, the command does nothing.
//
// On Windows, where SHELL is not commonly set, PATCH_ENV_COMMAND is passed
// to exec.Command() directly.
func Patch() error {
	cmdString := os.Getenv(patchCommandVar)
	if cmdString == "" {
		return nil
	}

	return patchFromCommand(cmdString)
}

// patchFromCommand runs the specified command string in the shell (if
// possible) and updates the running process's environment from its output.
func patchFromCommand(cmdString string) error {
	outBuf, err := runWithShell(cmdString)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(outBuf)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			log.Printf("[WARNING] patchenv: invalid output line: %s", line)
			continue
		}

		err := os.Setenv(parts[0], parts[1])
		if err != nil {
			log.Printf("[WARNING] patchenv: os.Setenv(%q, %q) returned error: %s",
				parts[0], parts[1], err)
		}
	}
	return nil
}

// runWithShell runs the specified command with the user's shell, as indicated
// by the SHELL environment variable.  The shell program is assumed to accept
// the POSIX "-c" command-line option.  If SHELL isn't set, the command string
// is passed as the first argument to exec.Command (on Windows SHELL usually
// isn't set, but programs parse their own command-line arguments, so this is
// the expected behavior there).
func runWithShell(cmdString string) (*bytes.Buffer, error) {
	var cmd *exec.Cmd

	shell := os.Getenv(shellVar)
	if shell == "" {
		cmd = exec.Command(cmdString)
	} else {
		cmd = exec.Command(shell, "-c", cmdString)
	}

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err := cmd.Run()
	if err != nil {
		_, _ = os.Stdout.Write(outBuf.Bytes())
		_, _ = os.Stderr.Write(errBuf.Bytes())
		return nil, fmt.Errorf("patchenv command %q failed: %q",
			cmdString, err.Error())
	}

	return outBuf, nil
}
