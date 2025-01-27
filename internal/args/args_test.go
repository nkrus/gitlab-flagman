package args

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name          string
		flags         map[string]string
		expectedError string
		expectedArgs  Args
	}{
		{
			name: "valid arguments",
			flags: map[string]string{
				"gitLabToken":     "token123",
				"gitLabProjectID": "123456",
			},
			expectedError: "",
			expectedArgs: Args{
				FlagsFile:            defaultFlagsFile,
				GitLabBase:           defaultGitLabBase,
				GitLabToken:          "token123",
				GitLabProjectID:      "123456",
				GitLabRequestTimeout: 10,
			},
		},
		{
			name: "missing gitLabToken",
			flags: map[string]string{
				"gitLabProjectID": "123456",
			},
			expectedError: "-gitLabToken обязателен",
		},
		{
			name: "missing gitLabProjectID",
			flags: map[string]string{
				"gitLabToken": "token123",
			},
			expectedError: "-gitLabProjectID обязателен",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetFlags(t)
			RegisterFlags()
			for key, value := range tc.flags {
				err := flag.Set(key, value)
				require.NoError(t, err)
			}

			parsedArgs, err := ParseArgs()

			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedArgs, *parsedArgs)
		})
	}
}

// resetFlags сбрасывает флаги между тестами
func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}
