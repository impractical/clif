package clif_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"impractical.co/clif"
)

func TestRoute(t *testing.T) {
	t.Parallel()
	type testCase struct {
		app             clif.Application
		input           []string
		expectedCmdName string
		expectedFlags   clif.FlagSet
		expectedArgs    []string
		expectedErr     error
	}

	cases := map[string]testCase{
		"basic": {
			input:           []string{"help"},
			app:             clif.Application{Commands: []clif.Command{{Name: "help", Handler: funcCommandHandler(func(_ context.Context, _ *clif.Response) {})}}},
			expectedCmdName: "help",
			expectedFlags:   clif.FlagSet{},
		},
		"list-flags": {
			input:           []string{"hello", "--name=foo", "--name", "bar", "--name", "baaz"},
			app:             clif.Application{Commands: []clif.Command{{Name: "hello", Flags: []clif.FlagDef{{Name: "--name", AllowMultiple: true}}, Handler: funcCommandHandler(func(_ context.Context, _ *clif.Response) {})}}},
			expectedCmdName: "hello",
			expectedFlags: clif.FlagSet{
				"name": {
					{HasValue: true, Raw: "foo", Key: "name", CanonicalKey: "name"},
					{HasValue: true, Raw: "bar", Key: "name", CanonicalKey: "name"},
					{HasValue: true, Raw: "baaz", Key: "name", CanonicalKey: "name"},
				},
			},
		},
		"mistaken-short-flags": {
			input:           []string{"hello", "-name=foo"},
			app:             clif.Application{Commands: []clif.Command{{Name: "hello", Flags: []clif.FlagDef{{Name: "--name"}}, Handler: funcCommandHandler(func(_ context.Context, _ *clif.Response) {})}}},
			expectedCmdName: "hello",
			expectedErr:     clif.ShortFlagsWithValueError{Flag: "-name=foo"},
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := clif.Route(context.Background(), testCase.app, testCase.input)
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
			if res.Command.Name != testCase.expectedCmdName {
				t.Errorf("Expected command %q, got %q", testCase.expectedCmdName, res.Command.Name)
			}
			if diff := cmp.Diff(testCase.expectedFlags, res.Flags); diff != "" {
				t.Errorf("Unexpected diff comparing flags (-expected, +got): %s", diff)
			}
			if diff := cmp.Diff(testCase.expectedArgs, res.Args); diff != "" {
				t.Errorf("Unexpected diff comparing args (-expected, +got): %s", diff)
			}
		})
	}
}
