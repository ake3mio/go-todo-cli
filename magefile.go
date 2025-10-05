//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Build() error {
	fmt.Println("Building...")

	if err := clean(); err != nil {
		return err
	}

	if err := execute("go build -o ./dist/todo main.go"); err != nil {
		return err
	}

	return nil
}

func clean() error {
	if err := execute("rm -fr ./dist"); err != nil {
		return err
	}
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return err
	}
	return nil
}

func Test() error {
	fmt.Println("Testing...")
	if err := execute("go test ./... -test.v"); err != nil {
		return err
	}

	return nil
}

func execute(command string) error {
	parts := strings.Fields(command)
	fmt.Printf("%v %v\n", parts[0], strings.Join(parts[1:], " "))
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var Default = Build
