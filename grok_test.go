package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/vjeantet/grok"
)

func TestAddDefaultPatterns(t *testing.T) {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		t.Fatalf("Grok engine initialization failed: %v\n", err)
	}

	err = AddDefaultPatterns(g)
	if err != nil {
		t.Error(err)
	}
}

func TestReadPatternsFromFile(t *testing.T) {
	temporaryDirectory, err := os.MkdirTemp(os.TempDir(), "sentlog-*")
	if err != nil {
		t.Fatalf("creating temporary directory: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(temporaryDirectory)
		if err != nil {
			t.Log(err)
		}
	})

	var tests = []struct {
		testName      string
		fileName      string
		payload       []byte
		expectedError error
	}{
		{
			testName: "Normal case",
			fileName: path.Join(temporaryDirectory, "normal.txt"),
			payload: []byte(`postgres ^%{DATESTAMP:timestamp}.*FATAL:.*host"

# This is a comment, should be skipped`),
			expectedError: nil,
		},
		{
			testName:      "Invalid length",
			fileName:      path.Join(temporaryDirectory, "invalid-length.txt"),
			payload:       []byte(`HelloWorld`),
			expectedError: fmt.Errorf("Cannot parse patterns in \"%s\"", path.Join(temporaryDirectory, "invalid-length.txt")),
		},
		{
			testName:      "Invalid pattern",
			fileName:      path.Join(temporaryDirectory, "invalid-pattern.txt"),
			payload:       []byte(`hello %{HELLO-WORLD}`),
			expectedError: errors.New("no pattern found for %{HELLO-WORLD}"),
		},
		{
			testName:      "Empty file",
			fileName:      path.Join(temporaryDirectory, "empty.txt"),
			payload:       nil,
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
			if err != nil {
				t.Fatalf("Grok engine initialization failed: %v\n", err)
			}

			err = os.WriteFile(test.fileName, test.payload, 0644)
			if err != nil {
				t.Fatalf("writing temporary file: %v", err)
			}

			err = ReadPatternsFromFile(g, test.fileName)
			if test.expectedError == nil && err != nil {
				t.Error(err)
			} else if test.expectedError != nil {
				if err != nil {
					if test.expectedError.Error() != err.Error() {
						t.Errorf("expecting %s, got %s", test.expectedError.Error(), err.Error())
					}
				} else {
					// err is nil
					t.Errorf("expecting %s, got nil", test.expectedError.Error())
				}
			}
		})
	}
}
