package clif

import (
	"context"
	"errors"
	"testing"
)

func TestCommand_parse(t *testing.T) {
	t.Parallel()
	type testCase struct {
		app               Application
		args              []string
		allowNonFlagFlags bool
		expectedCmdName   string
		//expectedFlags     []Flag
		//expectedArgs      []string
		//expectedUnparsed  []string
		expectedErr error
	}

	cases := map[string]testCase{
		"basic": {
			args:            []string{"help"},
			app:             Application{Commands: []Command{{Name: "help"}}},
			expectedCmdName: "help",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := parse(context.Background(), testCase.app, testCase.args, testCase.allowNonFlagFlags)
			if err != nil && testCase.expectedErr == nil {
				t.Fatalf("Unexpected error: %+v", err)
			}
			if err == nil && testCase.expectedErr != nil {
				t.Fatal("Expected error, didn't get one")
			}
			if err != nil && testCase.expectedErr != nil && !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("Expected error %+v, got error %+v", testCase.expectedErr, err)
			}
			if err != nil && testCase.expectedErr != nil {
				return
			}
			if res.subcommand.Name != testCase.expectedCmdName {
				t.Fatalf("Expected command %q, got %q", testCase.expectedCmdName, res.subcommand.Name)
			}
		})
	}
}
