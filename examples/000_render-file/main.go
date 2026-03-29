// Render a Markdown file to styled terminal output.
// Run: go run ./examples/000_render-file/ [path/to/file.md]
//
// If no file is given, it renders the bundled sample.md.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/indaco/herald"
	heraldmd "github.com/indaco/herald-md"
)

func main() {
	path := mdPath()

	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ty := herald.New()
	fmt.Println(heraldmd.Render(ty, source))
}

// mdPath returns the Markdown file to render: the first CLI argument,
// or the bundled sample.md next to this source file.
func mdPath() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	_, src, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(src), "sample.md")
}
