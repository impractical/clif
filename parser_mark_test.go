package clif

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser_mark(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected []*inputToken
	}

	testCases := map[string]testCase{
		"no subcommands args or flags": {
			input:    []string{},
			expected: []*inputToken{},
		},
		"subcommand or arg": {
			input: []string{"subcommand"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
			},
		},
		"long flag": {
			input: []string{"--flag-key"},
			expected: []*inputToken{
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"long flag and value": {
			input: []string{"--flag-key=flag-value"},
			expected: []*inputToken{
				{
					value:     "--flag-key=flag-value",
					tokenType: tokenTypeFlagKeyAndValue,
				},
			},
		},
		"long flag and value no equals": {
			input: []string{"--flag-key", "flag-value"},
			expected: []*inputToken{
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "flag-value",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeFlagValue: {},
						tokenTypeCommand:   {},
						tokenTypeArgument:  {},
					},
				},
			},
		},
		"short flag": {
			input: []string{"-a"},
			expected: []*inputToken{
				{
					value:     "-a",
					tokenType: tokenTypeShortFlag,
				},
			},
		},
		"multiple short flags": {
			input: []string{"-abc"},
			expected: []*inputToken{
				{
					value:     "-abc",
					tokenType: tokenTypeShortFlag,
				},
			},
		},
		"subcommand or arg then subcommand or arg": {
			input: []string{"subcommand", "subsubcommand"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "subsubcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
			},
		},
		"subcommand or arg then long flag": {
			input: []string{"subcommand", "--flag-key"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"subcommand or arg then long flag and value": {
			input: []string{"subcommand", "--flag-key=flag-value"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "--flag-key=flag-value",
					tokenType: tokenTypeFlagKeyAndValue,
				},
			},
		},
		"subcommand or arg then long flag and value no equals": {
			input: []string{"subcommand", "--flag-key", "flag-value"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "flag-value",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeFlagValue: {},
						tokenTypeCommand:   {},
						tokenTypeArgument:  {},
					},
				},
			},
		},
		"subcommand or arg then short flag": {
			input: []string{"subcommand", "-a"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "-a",
					tokenType: tokenTypeShortFlag,
				},
			},
		},
		"subcommand or arg then multiple short flags": {
			input: []string{"subcommand", "-abc"},
			expected: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
				{
					value:     "-abc",
					tokenType: tokenTypeShortFlag,
				},
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			parser := newParser(context.Background(), test.input)
			parser.mark(context.Background())
			if diff := cmp.Diff(test.expected, parser.tokens); diff != "" {
				t.Errorf("Unexpected results for %q (-expected, +got): %s", strings.Join(test.input, " "), diff)
			}
		})
	}
}
