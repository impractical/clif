package clif

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCommand_parse(t *testing.T) {
	t.Parallel()
	type testCase struct {
		app               Application
		args              []string
		allowNonFlagFlags bool
		expectedCmdName   string
		expectedFlags     map[string]Flag
		//expectedArgs      []string
		//expectedUnparsed  []string
		expectedErr error
	}

	cases := map[string]testCase{
		"basic": {
			args:            []string{"help"},
			app:             Application{Commands: []Command{{Name: "help"}}},
			expectedCmdName: "help",
			expectedFlags:   map[string]Flag{},
		},
		"list-flags": {
			args:            []string{"--name=foo", "--name", "bar", "--name", "baaz", "hello"},
			app:             Application{Commands: []Command{{Name: "hello", Flags: []FlagDef{{Name: "name", ValueAccepted: true, Parser: StringListParser{}}}}}},
			expectedCmdName: "hello",
			expectedFlags: map[string]Flag{
				"name": ListFlag[string]{Name: "name", RawValue: "foo, bar, baaz", Value: []string{"foo", "bar", "baaz"}},
			},
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
			if diff := cmp.Diff(testCase.expectedFlags, res.flags); diff != "" {
				t.Fatalf("Unexpected diff comparing flags (-expected, +got): %s", diff)
			}
		})
	}
}
