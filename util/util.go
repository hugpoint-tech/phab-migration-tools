package util

import (
	"fmt"
	"os"
)

func CheckFatal(msg string, e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "fatal %s %s\n", msg, e)
		panic(e)
	}
}
