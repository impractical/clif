package clif

import (
	"context"
	"fmt"
	"strings"
)

type commandPathEntry struct {
	cmd  Command
	name string
}

type inputRouter struct {
	cmd    Command
	flags  map[string][]*string
	args   []string
	path   []commandPathEntry
	parser *inputParser
}

func (router *inputRouter) route(ctx context.Context) error {
	for pos, token := range router.parser.tokens {
		switch {
		case token.isType(tokenTypeCommand):
			sub, err := router.getSubcommand(ctx, router.path, token.value)
			if err != nil {
				return err
			}
			router.cmd = sub
			router.path = append(router.path, commandPathEntry{
				cmd:  sub,
				name: token.value,
			})
		case token.isType(tokenTypeFlagKey):
			// if there's a next token *and* the next token is a
			// flag value, it's this flag's value
			if pos < len(router.parser.tokens)-1 && router.parser.tokens[pos+1].isType(tokenTypeFlagValue) {
				router.flags[token.value] = append(router.flags[token.value], &router.parser.tokens[pos+1].value)
				continue
			}
			// if there's not a next token or it's not a flag
			// value, record a nil to track that the flag was used
			// without a value
			router.flags[token.value] = append(router.flags[token.value], nil)
		case token.isType(tokenTypeArgument):
			router.args = append(router.args, token.value)
		case token.isType(tokenTypeUnknown):
			mightBe := []tokenType{}
			for k := range token.mightBe {
				mightBe = append(mightBe, k)
			}
			return UnknownTokenTypeError{
				value:   token.value,
				mightBe: mightBe,
			}
		case token.isType(tokenTypeFlagValue):
			// ignore the flag values, we've already taken care of
			// them when we saw the flag keys
			continue
		default:
			return UnexpectedTokenTypeError{
				tokenType: token.tokenType,
				value:     token.value,
			}
		}
	}
	return nil
}

func (router *inputRouter) getSubcommand(_ context.Context, path []commandPathEntry, command string) (Command, error) {
	for _, sub := range router.cmd.Subcommands {
		if sub.Name == command {
			return sub, nil
		}
		for _, alias := range sub.Aliases {
			if alias == command {
				return sub, nil
			}
		}
	}
	stringPath := make([]string, 0, len(path))
	for _, entry := range path {
		stringPath = append(stringPath, entry.name)
	}
	return Command{}, UnknownCommandError{Path: append(stringPath, command)}
}

// UnexpectedTokenTypeError is returned when the router encounters a token type
// it doesn't know how to handle. This usually indicates a bug in clif.
type UnexpectedTokenTypeError struct {
	// tokenType is the unhandled tokenType.
	tokenType tokenType

	// value is the value that got classified as an unhandled tokenType.
	value string
}

func (err UnexpectedTokenTypeError) Error() string {
	return fmt.Sprintf("unexpected token type %s for %s", err.tokenType, err.value)
}

// UnknownTokenTypeError is returned when the router encounters a token and
// isn't sure what type it is. This usually indicates a bug in clif.
type UnknownTokenTypeError struct {
	// mightBe are the different tokenTypes a token could be.
	mightBe []tokenType

	// value is the value that could be multiple token types.
	value string
}

func (err UnknownTokenTypeError) Error() string {
	mightBe := []string{}
	for _, k := range err.mightBe {
		mightBe = append(mightBe, k.String())
	}
	return fmt.Sprintf("unknown token type for %q, could be: %s", err.value, strings.Join(mightBe, ", "))
}
