package cli

import (
	"testing"
)

func TestParseSingleFlags(t *testing.T) {
	tests := []struct {
		args    []string
		check   func(*Options) bool
		desc    string
	}{
		{[]string{"-f"}, func(o *Options) bool { return o.Force }, "force flag"},
		{[]string{"-r"}, func(o *Options) bool { return o.Recursive }, "recursive lowercase"},
		{[]string{"-R"}, func(o *Options) bool { return o.Recursive }, "recursive uppercase"},
		{[]string{"-i"}, func(o *Options) bool { return o.Interactive }, "interactive flag"},
		{[]string{"-v"}, func(o *Options) bool { return o.Verbose }, "verbose flag"},
		{[]string{"-d"}, func(o *Options) bool { return o.RemoveEmptyDirs }, "remove empty dirs"},
		{[]string{"--force"}, func(o *Options) bool { return o.Force }, "force long"},
		{[]string{"--recursive"}, func(o *Options) bool { return o.Recursive }, "recursive long"},
		{[]string{"--verbose"}, func(o *Options) bool { return o.Verbose }, "verbose long"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !tt.check(opts) {
				t.Errorf("Parse(%v) flag not set correctly", tt.args)
			}
		})
	}
}

func TestParseCombinedFlags(t *testing.T) {
	tests := []struct {
		args     []string
		wantF    bool
		wantR    bool
		wantV    bool
		desc     string
	}{
		{[]string{"-rf"}, true, true, false, "combined rf"},
		{[]string{"-fr"}, true, true, false, "combined fr"},
		{[]string{"-rfv"}, true, true, true, "combined rfv"},
		{[]string{"-Rf"}, true, true, false, "combined Rf uppercase"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if opts.Force != tt.wantF {
				t.Errorf("Force = %v, want %v", opts.Force, tt.wantF)
			}
			if opts.Recursive != tt.wantR {
				t.Errorf("Recursive = %v, want %v", opts.Recursive, tt.wantR)
			}
			if opts.Verbose != tt.wantV {
				t.Errorf("Verbose = %v, want %v", opts.Verbose, tt.wantV)
			}
		})
	}
}

func TestParseFiles(t *testing.T) {
	tests := []struct {
		args      []string
		wantFiles []string
		desc      string
	}{
		{[]string{"file1.txt"}, []string{"file1.txt"}, "single file"},
		{[]string{"file1.txt", "file2.txt"}, []string{"file1.txt", "file2.txt"}, "multiple files"},
		{[]string{"-f", "file.txt"}, []string{"file.txt"}, "flag then file"},
		{[]string{"file.txt", "-f"}, []string{"file.txt"}, "file then flag"},
		{[]string{"--", "-f"}, []string{"-f"}, "double dash escapes flag"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(opts.Files) != len(tt.wantFiles) {
				t.Fatalf("Files count = %d, want %d", len(opts.Files), len(tt.wantFiles))
			}
			for i, f := range opts.Files {
				if f != tt.wantFiles[i] {
					t.Errorf("Files[%d] = %q, want %q", i, f, tt.wantFiles[i])
				}
			}
		})
	}
}

func TestParseSafeRmFlags(t *testing.T) {
	tests := []struct {
		args    []string
		check   func(*Options) bool
		desc    string
	}{
		{[]string{"--safe-list"}, func(o *Options) bool { return o.SafeList }, "safe list"},
		{[]string{"--safe-restore=/path"}, func(o *Options) bool { return o.SafeRestore == "/path" }, "safe restore"},
		{[]string{"--safe-purge"}, func(o *Options) bool { return o.SafePurge }, "safe purge"},
		{[]string{"--purge-days=7"}, func(o *Options) bool { return o.PurgeDays == 7 }, "purge days"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !tt.check(opts) {
				t.Errorf("Parse(%v) safe-rm flag not set correctly", tt.args)
			}
		})
	}
}

func TestParsePreserveRoot(t *testing.T) {
	// Default should be preserve root
	opts, _ := Parse([]string{})
	if !opts.PreserveRoot {
		t.Error("Default should have PreserveRoot = true")
	}

	// --no-preserve-root should override
	opts, _ = Parse([]string{"--no-preserve-root"})
	if opts.PreserveRoot {
		t.Error("--no-preserve-root should set PreserveRoot = false")
	}
	if !opts.NoPreserveRoot {
		t.Error("--no-preserve-root should set NoPreserveRoot = true")
	}
}

func TestParseInvalidFlag(t *testing.T) {
	_, err := Parse([]string{"-x"})
	if err == nil {
		t.Error("Parse should return error for invalid flag")
	}
}
