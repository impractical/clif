package clif

import (
	"context"
	"fmt"
	"strings"
)

type tokenType uint8

const (
	tokenTypeUnknown         tokenType = 0
	tokenTypeFlagKey         tokenType = 1
	tokenTypeFlagValue       tokenType = 2
	tokenTypeCommand         tokenType = 3
	tokenTypeArgument        tokenType = 4
	tokenTypeFlagKeyAndValue tokenType = 5
	tokenTypeShortFlag       tokenType = 6
)

func (t tokenType) String() string {
	switch t {
	case tokenTypeUnknown:
		return "unknown"
	case tokenTypeFlagKey:
		return "flagKey"
	case tokenTypeFlagValue:
		return "flagValue"
	case tokenTypeCommand:
		return "command"
	case tokenTypeArgument:
		return "argument"
	case tokenTypeFlagKeyAndValue:
		return "flagKeyAndValue"
	case tokenTypeShortFlag:
		return "shortFlag"
	default:
		return fmt.Sprintf("invalid(%d)", t)
	}
}

type inputParser struct {
	tokens []*inputToken
}

type inputToken struct {
	value     string
	tokenType tokenType
	mightBe   map[tokenType]struct{}
}

func (token inputToken) isType(candidateType tokenType) bool {
	return token.tokenType == candidateType
}

func (token inputToken) canBeType(candidateType tokenType) bool {
	if token.isType(candidateType) {
		return true
	}
	if !token.isType(tokenTypeUnknown) {
		return false
	}
	_, ok := token.mightBe[candidateType]
	return ok
}

func (token *inputToken) setPotentialType(potentialType tokenType) {
	if token.mightBe == nil {
		token.mightBe = map[tokenType]struct{}{
			potentialType: {},
		}
		return
	}
	token.mightBe[potentialType] = struct{}{}
}

func (token *inputToken) removePotentialType(potentialType tokenType) {
	if token.mightBe == nil {
		return
	}
	delete(token.mightBe, potentialType)
}

func (token *inputToken) setType(definiteType tokenType) {
	token.tokenType = definiteType
	token.mightBe = map[tokenType]struct{}{}
}

// Equal returns whether `token` is semantically equivalent to `other` or not.
func (token *inputToken) Equal(other *inputToken) bool {
	if token == nil && other == nil {
		return true
	}
	if token == nil && other != nil {
		return false
	}
	if token != nil && other == nil {
		return false
	}
	if token.value != other.value {
		return false
	}
	if token.tokenType != other.tokenType {
		return false
	}
	if len(token.mightBe) != len(other.mightBe) {
		return false
	}
	for k := range token.mightBe {
		if _, ok := other.mightBe[k]; !ok {
			return false
		}
	}
	return true
}

func newParser(_ context.Context, inputs []string) *inputParser {
	parser := &inputParser{
		tokens: []*inputToken{},
	}
	for _, input := range inputs {
		parser.tokens = append(parser.tokens, &inputToken{value: input})
	}
	return parser
}

func (parser *inputParser) mark(_ context.Context) {
	prevWasFlagKey := false
	postSeparator := false
	for _, token := range parser.tokens {
		if postSeparator {
			token.tokenType = tokenTypeArgument
			continue
		}
		if token.value == "--" {
			token.tokenType = tokenTypeArgument
			prevWasFlagKey = false
			postSeparator = true
			continue
		}
		if strings.HasPrefix(token.value, "--") {
			if strings.Contains(token.value, "=") {
				token.tokenType = tokenTypeFlagKeyAndValue
				prevWasFlagKey = false
				continue
			}
			token.tokenType = tokenTypeFlagKey
			prevWasFlagKey = true
			continue
		}
		if strings.HasPrefix(token.value, "-") && token.value != "-" {
			token.tokenType = tokenTypeShortFlag
			prevWasFlagKey = false
			continue
		}
		if prevWasFlagKey {
			token.setPotentialType(tokenTypeFlagValue)
		}
		token.setPotentialType(tokenTypeArgument)
		token.setPotentialType(tokenTypeCommand)
		prevWasFlagKey = false
	}
}

func (parser *inputParser) normalize(_ context.Context) {
	tokens := []*inputToken{}
	for _, token := range parser.tokens {
		switch token.tokenType { //nolint:exhaustive // we have a default defined for a reason
		case tokenTypeFlagKeyAndValue:
			key, value, _ := strings.Cut(token.value, "=")
			tokens = append(tokens, &inputToken{
				value:     key,
				tokenType: tokenTypeFlagKey,
			}, &inputToken{
				value:     value,
				tokenType: tokenTypeFlagValue,
			})
		case tokenTypeShortFlag:
			pieces := strings.Split(token.value, "")
			// first piece is just a hyphen
			pieces = pieces[1:]
			for _, piece := range pieces {
				tokens = append(tokens, &inputToken{
					value:     "-" + piece,
					tokenType: tokenTypeFlagKey,
				})
			}
		case tokenTypeUnknown:
			// if there's only one thing it might be, we know what
			// it is
			if len(token.mightBe) == 1 {
				var typ tokenType
				for k := range token.mightBe {
					typ = k
				}
				tokens = append(tokens, &inputToken{
					value:     token.value,
					tokenType: typ,
				})
			} else {
				tok := *token
				tokens = append(tokens, &tok)
			}
		default:
			tok := *token
			tokens = append(tokens, &tok)
		}
	}
	parser.tokens = tokens
}

type inputParserCommand struct {
	subcommands map[string]inputParserCommand
	allowsArgs  bool
	flags       map[string]inputParserFlag
}

type inputParserFlag struct {
	allowsValue   bool
	mustHaveValue bool
}

func (parser *inputParser) apply(ctx context.Context, cmd inputParserCommand) error {
	possibleFlags := map[string]inputParserFlag{}
	for k, v := range cmd.flags {
		possibleFlags[k] = v
	}
	var cmdPath []string
	for pos, token := range parser.tokens {
		// subcommands are easy; if we match a subcommand defined on
		// this command, we assume we want that subcommand
		if token.canBeType(tokenTypeCommand) {
			subcmd, ok := cmd.subcommands[token.value]
			if ok {
				// if this matches a subcommand, assume it's
				// that subcommand and update the state we're
				// working with, including the possible flags
				cmd = subcmd
				for key, flag := range subcmd.flags {
					if _, ok := possibleFlags[key]; ok {
						// reused a flag in the same branch, throw error
						return DuplicateFlagNameError(key)
					}
					possibleFlags[key] = flag
				}
				token.setType(tokenTypeCommand)
				cmdPath = append(cmdPath, token.value)
				continue
			}
			if token.isType(tokenTypeCommand) {
				// we know this is a command, but we don't have a definition for it
				return UnknownCommandError{Path: append(cmdPath, token.value)}
			}

			// if we aren't definitely a command, we're just not
			// maybe a command, but could be something else
			token.removePotentialType(tokenTypeCommand)
		}

		// if we know what we are, there's no need to keep going. We
		// needed to run the command block so state is updated, but we
		// don't need to mess with the other stuff if we know what we
		// are.
		if !token.isType(tokenTypeUnknown) {
			continue
		}

		// arguments are also easy; if we don't allow arguments on this
		// command, it can't be an argument
		if token.canBeType(tokenTypeArgument) && !cmd.allowsArgs {
			token.removePotentialType(tokenTypeArgument)
		}

		// flag values are a bit tougher; we need to know whether the
		// flag key preceding it accepts arguments or not
		if token.canBeType(tokenTypeFlagValue) && pos > 0 && parser.tokens[pos-1].isType(tokenTypeFlagKey) {
			flag, ok := possibleFlags[parser.tokens[pos-1].value]
			if ok {
				if flag.allowsValue {
					// assume if the flag wants a value and
					// this isn't a command, it's the flag
					// value
					token.setType(tokenTypeFlagValue)
					continue
				}
				// if the flag can't have a value, this can't
				// be a value
				token.removePotentialType(tokenTypeFlagValue)
			}
			// if we didn't find the flag, it's too early to say
			// that it's an invalid flag; it could be defined in a
			// future subcommand
		}
	}

	// normalize, so tokens that can only be one thing are updated to be
	// that one thing
	parser.normalize(ctx)

	// loop through again looking for flags defined on future subcommands,
	// now that we presumably have all flags from all subcommands parsed
	for pos, token := range parser.tokens {
		if pos == 0 && token.canBeType(tokenTypeFlagValue) {
			// the first token can't be a flag value; flag values
			// must come after flag keys
			token.removePotentialType(tokenTypeFlagValue)
		}
		if !token.isType(tokenTypeUnknown) || !token.canBeType(tokenTypeFlagValue) {
			continue
		}

		if !parser.tokens[pos-1].isType(tokenTypeFlagKey) {
			// redundant, but may as well check to be sure
			return FlagValueWithoutFlagKeyError{Value: token.value}
		}

		flag, ok := possibleFlags[parser.tokens[pos-1].value]
		if !ok {
			// *now* we can say this is a problem, we know all
			// valid flags, and this isn't a valid flag
			return UnknownFlagNameError(parser.tokens[pos-1].value)
		}

		if flag.allowsValue {
			// assume if the flag wants a value and this isn't a
			// command, it's the flag value
			token.setType(tokenTypeFlagValue)
			continue
		}

		// if the flag can't have a value, this
		// can't be a value
		token.removePotentialType(tokenTypeFlagValue)
	}

	// normalize again
	parser.normalize(ctx)

	return nil
}
