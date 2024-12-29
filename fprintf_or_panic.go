package clif

import (
	"fmt"
	"io"
)

func fprintOrPanic(w io.Writer, a ...any) {
	_, err := fmt.Fprint(w, a...)
	if err != nil {
		panic(fmt.Sprintf("Error writing response: %s", err))
	}
}

func fprintlnOrPanic(w io.Writer, a ...any) {
	_, err := fmt.Fprintln(w, a...)
	if err != nil {
		panic(fmt.Sprintf("Error writing response: %s", err))
	}
}
