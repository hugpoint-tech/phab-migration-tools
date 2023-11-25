package util

import (
	"fmt"
	"os"
)

func CheckFatal(e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %s", e)
		os.Exit(1)
	}
}
