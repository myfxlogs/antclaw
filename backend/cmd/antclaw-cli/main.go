// Package main implements the AntClaw CLI tool.
//
// Usage:
//   antclaw-cli rotate-master-key --old=<base64> --new=<base64> [--dry-run]
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: antclaw-cli <command> [args...]")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  rotate-master-key    Rotate master key for encrypted secrets")
		os.Exit(1)
	}

	var exitCode int
	switch os.Args[1] {
	case "rotate-master-key":
		exitCode = rotateMasterKeyCmd(os.Args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	os.Exit(exitCode)
}
