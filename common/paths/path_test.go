package paths

import "testing"

func TestFileAndExtension(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedExt  string
	}{
		{"path/to/file.md", "file", ".md"},
		{"file", "file", ""},
		{"path/to/", "", ""},
		{"", "", ""},
		{"path/to/file.tar.gz", "file.tar", ".gz"},
		{"path/to/.hiddenfile", "", ".hiddenfile"},
	}

	for _, tt := range tests {
		gotName, gotExt := FileAndExtension(tt.input)
		if gotName != tt.expectedName || gotExt != tt.expectedExt {
			t.Errorf("FileAndExtension(%q) = %q, %q; want %q, %q | sep: %q",
				tt.input, gotName, gotExt, tt.expectedName, tt.expectedExt, PathSeparator)
		}
	}
}

func TestFileAndExtensionNoDelimiter(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedExt  string
	}{
		{"path/to/file.md", "file", "md"},
		{"file", "file", ""},
		{"path/to/", "", ""},
		{"", "", ""},
		{"path/to/file.tar.gz", "file.tar", "gz"},
		{"path/to/.hiddenfile", "", "hiddenfile"},
	}

	for _, tt := range tests {
		gotName, gotExt := FileAndExtensionNoDelimiter(tt.input)
		if gotName != tt.expectedName || gotExt != tt.expectedExt {
			t.Errorf("FileAndExtension(%q) = %q, %q; want %q, %q",
				tt.input, gotName, gotExt, tt.expectedName, tt.expectedExt)
		}
	}
}

func TestSplitDirFile(t *testing.T) {
	tests := []struct {
		inputPath string
		expected  DirFile
	}{
		{"path/to/file.md", DirFile{"path/to/", "file.md"}},
		{"file", DirFile{"", "file"}},
		{"path/to/", DirFile{"path/to/", ""}},
		{"", DirFile{"", ""}},
	}

	for _, tt := range tests {
		got := SplitDirFile(tt.inputPath)
		if got != tt.expected {
			t.Errorf("SplitDirFile(%q) = %q; want %q",
				tt.inputPath, got, tt.expected)
		}
	}
}
