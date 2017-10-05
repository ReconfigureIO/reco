package printer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/ts"
)

// Table is the table to print.
type Table struct {
	Header []string   // Column titles
	Body   [][]string // Rows data
}

// Empty checks if the table is empty.
func (t Table) Empty() bool {
	return len(t.Body) == 0
}

// Print prints the table to stdout.
// It pipes to less (or more for windows) if
// height of table is higher than terminal's.
func Print(table Table) error {
	var buf bytes.Buffer
	Fprint(&buf, table)

	// attempt to get terminal size
	size, err := ts.GetSize()

	// if table height is less than terminal height,
	// print normally.
	if err != nil || len(table.Body) < size.Row()-2 {
		_, err := buf.WriteTo(os.Stdout)
		return err
	}

	// otherwise pipe to less (or more for windows).
	command := "less"
	if runtime.GOOS == "windows" {
		command = "more"
	}

	// if less (or more) is not found, print normally.
	if _, err := exec.LookPath(command); err != nil {
		_, err := buf.WriteTo(os.Stdout)
		return err
	}

	cmd := exec.Command(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = &buf
	return cmd.Run()
}

// Fprint prints the table to a writer.
func Fprint(w io.Writer, table Table) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	tWriter := tablewriter.NewWriter(w)
	tWriter.SetHeader(table.Header)
	tWriter.SetHeaderLine(false)
	tWriter.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	tWriter.SetCenterSeparator(" ")
	tWriter.SetColumnSeparator("    ")
	tWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	tWriter.AppendBulk(table.Body) // Add Bulk Data
	tWriter.Render()
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}
