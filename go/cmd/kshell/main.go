package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"text/template"

	"gopkg.in/yaml.v3"
)

func main() {
	var buf bytes.Buffer
	var data any

	dir := os.Getenv("SB_HOME")
	path := path.Join(dir, "pod-overrides.yaml")

	tpl, _ := template.ParseFiles(path)
	tpl.Execute(&buf, map[string]string{})

	yaml.Unmarshal(buf.Bytes(), &data)
	overrides, _ := json.Marshal(data)

	registry := os.Getenv("SB_REGISTRY")
	image := fmt.Sprintf("%s/%s", registry, os.Args[1])

	args := []string{"kubectl", "run", "kshell", "--rm", "-it", "--image", image, "--overrides", string(overrides)}

	kubectl, _ := exec.LookPath("kubectl")

	err := syscall.Exec(kubectl, args, os.Environ())
	if err != nil {
		panic(err)
	}
}
