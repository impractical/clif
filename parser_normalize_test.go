package clif

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser_normalize(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*inputToken
		expected []*inputToken
	}

	testCases := map[string]testCase{
		"no subcommands args or flags": {
			input:    []*inputToken{},
			expected: []*inputToken{},
		},
		"subcommand or arg": {
			input: []*inputToken{
				{
					value:     "subcommand",
					tokenType: tokenTypeUnknown,
					mightBe: map[tokenType]struct{}{
						tokenTypeCommand:  {},
						tokenTypeArgument: {},
					},
				},
			},
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
			input: []*inputToken{
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
			},
			expected: []*inputToken{
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"long flag and value": {
			input: []*inputToken{
				{
					value:     "--flag-key=flag-value",
					tokenType: tokenTypeFlagKeyAndValue,
				},
			},
			expected: []*inputToken{
				{
					value:     "--flag-key",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "flag-value",
					tokenType: tokenTypeFlagValue,
				},
			},
		},
		"long flag and value no equals": {
			input: []*inputToken{
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
			input: []*inputToken{
				{
					value:     "-a",
					tokenType: tokenTypeShortFlag,
				},
			},
			expected: []*inputToken{
				{
					value:     "-a",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"multiple short flags": {
			input: []*inputToken{
				{
					value:     "-abc",
					tokenType: tokenTypeShortFlag,
				},
			},
			expected: []*inputToken{
				{
					value:     "-a",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "-b",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "-c",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"subcommand or arg then subcommand or arg": {
			input: []*inputToken{
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
			input: []*inputToken{
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
			input: []*inputToken{
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
					tokenType: tokenTypeFlagValue,
				},
			},
		},
		"subcommand or arg then long flag and value no equals": {
			input: []*inputToken{
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
			input: []*inputToken{
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
					tokenType: tokenTypeFlagKey,
				},
			},
		},
		"subcommand or arg then multiple short flags": {
			input: []*inputToken{
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
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "-b",
					tokenType: tokenTypeFlagKey,
				},
				{
					value:     "-c",
					tokenType: tokenTypeFlagKey,
				},
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			parser := &inputParser{
				tokens: test.input,
			}
			parser.normalize(context.Background())
			input := []string{}
			for _, token := range test.input {
				input = append(input, token.value)
			}
			if diff := cmp.Diff(test.expected, parser.tokens); diff != "" {
				t.Errorf("Unexpected results for %q (-expected, +got): %s", strings.Join(input, " "), diff)
			}
		})
	}
}
