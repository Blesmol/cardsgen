package main

import (
	"fmt"
	"os"
	"os/exec"
)

// checkTools verifies that the external binaries required for PDF generation
// are available on PATH.
func checkTools(needWeasyprint bool) error {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return fmt.Errorf("pandoc not found on PATH; install it from https://pandoc.org and try again")
	}
	if needWeasyprint {
		if _, err := exec.LookPath("weasyprint"); err != nil {
			return fmt.Errorf("weasyprint not found on PATH; install it from https://weasyprint.org and try again")
		}
	}
	return nil
}

// runPandocPDF renders the combined markdown to a PDF using weasyprint.
func runPandocPDF(mdPath, cssPath, outPath string) error {
	args := []string{
		mdPath,
		"--from", "markdown",
		"--standalone",
		"--pdf-engine", "weasyprint",
		"--css", cssPath,
		"--metadata", "pagetitle=Cards",
		"--output", outPath,
	}
	return runPandoc(args)
}

// runPandocHTML renders the combined markdown to a standalone HTML file for
// debugging the layout in a browser.
func runPandocHTML(mdPath, cssPath, outPath string) error {
	args := []string{
		mdPath,
		"--from", "markdown",
		"--to", "html5",
		"--standalone",
		"--css", cssPath,
		"--metadata", "pagetitle=Cards",
		"--output", outPath,
	}
	return runPandoc(args)
}

// runPandoc executes pandoc with the given arguments, streaming its output.
func runPandoc(args []string) error {
	cmd := exec.Command("pandoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc failed: %w", err)
	}
	return nil
}
