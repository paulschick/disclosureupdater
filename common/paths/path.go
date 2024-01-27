package paths

import (
	"path/filepath"
	"strings"
)

var PathSeparator = string(filepath.Separator)

// FileAndExtension returns the file and extension of a given path.
func FileAndExtension(input string) (name string, ext string) {
	return fileAndExt(input, PathSeparator)
}

// FileAndExtensionNoDelimiter returns the file and extension of a given path, excluding the delimiter.
// If FileAndExtension returns ".md", FileAndExtensionNoDelimiter would return "md".
func FileAndExtensionNoDelimiter(input string) (name string, ext string) {
	file, ext := fileAndExt(input, PathSeparator)
	return file, strings.TrimPrefix(ext, ".")
}

// fileAndExt internal function to return the file and extension of a given path.
func fileAndExt(input, sep string) (name string, ext string) {
	if input == "" {
		return
	}

	ext = filepath.Ext(input)
	base := filepath.Base(input)

	return extractFilename(input, ext, base, PathSeparator), ext
}

// extractFilename
// source: github.com/gohugoio/hugo/common/paths/path.go
func extractFilename(in, ext, base, pathSeparator string) (name string) {
	// No file name cases. These are defined as:
	// 1. any "in" path that ends in a pathSeparator
	// 2. any "base" consisting of just an pathSeparator
	// 3. any "base" consisting of just an empty string
	// 4. any "base" consisting of just the current directory i.e. "."
	// 5. any "base" consisting of just the parent directory i.e. ".."
	if (strings.LastIndex(in, pathSeparator) == len(in)-1) || base == "" || base == "." || base == ".." || base == pathSeparator {
		name = "" // there is NO filename
	} else if ext != "" { // there was an Extension
		// return the filename minus the extension (and the ".")
		name = base[:strings.LastIndex(base, ".")]
	} else {
		// no extension case so just return base, which will
		// be the filename
		name = base
	}
	return
}

// DirFile holds the path to the directory and the file
type DirFile struct {
	Dir  string
	File string
}

// SplitDirFile splits a path into a directory and a file
func SplitDirFile(path string) DirFile {
	dir, file := filepath.Split(path)
	return DirFile{Dir: dir, File: file}
}
