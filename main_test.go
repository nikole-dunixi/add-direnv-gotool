package main

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleEnvRC(t *testing.T) {

	tcs := []string{
		"empty",
		"present-single",
		"present-single-with-nl",
		"present-multiline",
		"present-multiline-with-nl",
		"absent-single",
		"absent-single-with-nl",
		"absent-multiline",
		"absent-multiline-with-nl",
	}
	for testcase := range slices.Values(tcs) {
		tmpDir, err := os.MkdirTemp("", "add-direnv-gotool_test")
		err = os.MkdirAll(tmpDir, UnixDirectoryPermissions)
		require.NoError(t, err)
		t.Run(testcase, func(t *testing.T) {
			err = os.MkdirAll(filepath.Join(tmpDir, "envrc"), UnixDirectoryPermissions)
			require.NoError(t, err)
			inputFilename := filepath.Join("testdata", "envrc", testcase+".input")
			tempFilename := filepath.Join(tmpDir, "envrc", testcase+".input")
			outputFilename := filepath.Join("testdata", "envrc", testcase+".output")
			// Copy the input file to a temporary file that may
			// be modified freely
			{
				tmpFile, err := os.Create(tempFilename)
				require.NoError(t, err)
				inputFile, err := os.Open(inputFilename)
				require.NoError(t, err)
				_, err = io.Copy(tmpFile, inputFile)
				require.NoError(t, err)
				inputFile.Close()
				tmpFile.Close()
			}

			// Perform the opperation
			err = HandleEnvRC(
				tempFilename,
				"add-direnv-gotool",
				".tools-subdirectory",
			)
			require.NoError(t, err)
			// Validate the changed file against the expected output
			// Load the files, which should never have a problem
			actualFile, err := os.Open(tempFilename)
			require.NoError(t, err, "should be able to read testdata without issue")
			actualBytes, err := io.ReadAll(actualFile)
			require.NoError(t, err, "should be able to read testdata without issue")
			expectedFile, err := os.Open(outputFilename)
			require.NoError(t, err, "should be able to read testdata without issue")
			expectedBytes, err := io.ReadAll(expectedFile)
			require.NoError(t, err, "should be able to read testdata without issue")
			// Compare the two files to ensure the correct operations occurred
			require.Len(t, string(actualBytes), len(expectedBytes),
				"the file contained:\n%s\n\ninstead of:\n%s\n",
				string(actualBytes),
				string(expectedBytes),
			)
			for i := range actualBytes {
				actualByte := actualBytes[i]
				expectedByte := expectedBytes[i]
				require.Equal(t, expectedByte, actualByte,
					"the file contained:\n%s\n\ninstead of:\n%s\n",
					string(actualBytes),
					string(expectedBytes),
				)
			}
		})
	}
}
