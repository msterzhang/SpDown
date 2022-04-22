// Package utils /*
package utils

import (
	"fmt"
	"github.com/fatih/color"
)

func Info(format string, args ...interface{}) {
	fmt.Printf(color.BlueString("[INFO] ")+format+"\n", args...)
}

func Err(format string, args ...interface{}) {
	fmt.Printf(color.RedString("[ERRO] ")+format+"\n", args...)
}
