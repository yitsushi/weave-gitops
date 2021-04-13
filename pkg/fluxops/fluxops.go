package fluxops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/version"
)

var (
	fluxHandler = defaultFluxHandler
	fluxBinary  string
)

func FluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func FluxBinaryPath() string {
	return fluxBinary
}

func SetFluxHandler(f func(string) ([]byte, error)) {
	fluxHandler = f
}

func CallFlux(arglist string) ([]byte, error) {
	return fluxHandler(arglist)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	initFluxBinary()
	return CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func CallCommand(cmdstr string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	return cmd.CombinedOutput()
}

func CallCommandSeparatingOutputStreams(cmdstr string) ([]byte, []byte, error) {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func CallCommandForEffect(cmdstr string) error {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initFluxBinary() {
	if fluxBinary == "" {
		fluxPath, err := FluxPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve wego executable path: %v", err)
			os.Exit(1)
		}
		fluxBinary = fluxPath
	}
}

func Escape(cmd string) string {
	return strings.ReplaceAll(cmd, "'", "'\"'\"'")
}
