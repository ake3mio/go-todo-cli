//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	app    = env("APP", "todo")
	pkg    = env("PKG", "main.go")
	cgo    = env("CGO_ENABLED", "1")
	vers   = env("VERSION", "dev")
	commit = env("COMMIT", shortCommit())
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

func Release() error {
	if vers == "dev" {
		vers = time.Now().UTC().Format("20060102-150405")
	}
	if err := Cross(); err != nil {
		return err
	}
	return Checksums()
}

func Cross() error {
	plats := []string{
		"linux/amd64",
		"linux/arm64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
	}

	if err := clean(); err != nil {
		return err
	}

	for _, p := range plats {
		dest := filepath.Join("dist", strings.ReplaceAll(p, "/", "_"))
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		args := []string{
			"docker",
			"buildx",
			"build",
			"-f",
			"build.dockerfile",
			"--progress=plain",
			"--platform", p,
			"--target", "artifact",
			"--build-arg", "APP=" + app,
			"--build-arg", "PKG=" + pkg,
			"--build-arg", "CGO_ENABLED=" + cgo,
			"--build-arg", "VERSION=" + vers,
			"--build-arg", "COMMIT=" + commit,
			"--output", "type=local,dest=" + dest,
			".",
		}
		if err := execute(strings.Join(args, " ")); err != nil {
			return err
		}
	}
	return nil
}

func Checksums() error {
	f, err := os.Create("dist/SHA256SUMS.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	entries, err := os.ReadDir("dist")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join("dist", e.Name())
		files, _ := os.ReadDir(dir)
		for _, fi := range files {
			path := filepath.Join(dir, fi.Name())
			sum := sha256File(path)
			fmt.Fprintf(f, "%s  %s\n", sum, path)
		}
	}
	fmt.Println(">> wrote dist/SHA256SUMS.txt")
	return nil
}

func sha256File(path string) string {
	out, _ := exec.Command("sh", "-c", "sha256sum "+shellQuote(path)+" | awk '{print $1}' || shasum -a 256 "+shellQuote(path)+" | awk '{print $1}'").Output()
	return strings.TrimSpace(string(out))
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shortCommit() string {
	out, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "unknown"
	}
	return s
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func execute(command string) error {
	parts := strings.Fields(command)
	fmt.Printf("%v %v\n", parts[0], parts[1:])
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var Default = Build
