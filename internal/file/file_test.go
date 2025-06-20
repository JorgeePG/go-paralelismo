package file

import (
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid file",
			filePath: "test.txt",
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "Non-existent file",
			filePath: "nonexistent.txt",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Valid file" {
				err := os.WriteFile(tt.filePath, []byte(tt.expected), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.filePath) // Clean up after test
			}

			fh := FileHandler{}
			got, err := fh.ReadFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ReadFile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		content  string
		wantErr  bool
	}{
		{
			name:     "Write to file",
			filePath: "output.txt",
			content:  "Hello World",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fh := FileHandler{}
			err := fh.WriteFile(tt.filePath, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify the content
			got, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to read written file: %v", err)
			}
			if string(got) != tt.content {
				t.Errorf("WriteFile() content = %v, want %v", string(got), tt.content)
			}

			// Clean up
			os.Remove(tt.filePath)
		})
	}
}
