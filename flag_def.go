package clif

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// FlagDef describes the definition of a flag. [Application]s and [Command]s
// will use FlagDefs to describe what flags they accept.
type FlagDef struct {
	// Name is the name of the flag. It's what will be surfaced in
	// documentation and what the user will use when applying the flag to a
	// command. Names must be unique across the command and all parent or
	// child commands, or the parser won't know which command to apply the
	// flag to.
	Name string

	// Aliases holds any alternative names the flag should accept from the
	// user. Aliases are not surfaced in documentation, by default. Aliases
	// must be unique across all other aliases and names for the command
	// and all parent or child commands, or the parser won't know which
	// command to apply the flag to.
	Aliases []string

	// Description is a user-friendly description of what the flag does and
	// what it's for, to be presented as part of help output.
	Description string

	// OnlyToggle indicates whether or not the flag should allow a value.
	// If set to true, attempting to pass a value will surface an error.
	//
	// If OnlyToggle is true, IsToggle must also be true.
	OnlyToggle bool

	// IsToggle indicates whether or not the flag should require a value.
	// If set to false, attempting to not pass a value with surface an
	// error.
	IsToggle bool

	// AllowMultiple indicates whether multiple instances of this flag
	// should be accepted, with the values returned as a list. False means
	// that specifying the flag more than once will result in an error.
	AllowMultiple bool

	// FromEnvVars indicates which, if any, environment variables should be
	// used as default values if the flag isn't specified. FromEnvVars
	// should be set to the list of environment variables to check, and the
	// first one to hold a non-empty value will be used as the flag value.
	//
	// It's important to note that the calling application will not be able
	// to distinguish between a flag value set via environment variable and
	// via actual flag; they'll appear the same. If you need to distinguish
	// between those cases, do not use FromEnvVars, and check the
	// environment variable(s) yourself in the Build method of your
	// [HandlerBuilder].
	FromEnvVars []string

	// Default sets the default values of this flag if it's not specified.
	// If FromEnvVars is set, the flag must not be specified in the
	// invocation or in the environment variables before the default will
	// be used.
	//
	// It's important to note that the calling application will not be able
	// to distinguish between a flag value set via this default value and
	// via actual flag; they'll appear the same. If you need to distinguish
	// between those cases, do not use Default, and instead set the value
	// yourself in the Build method of your [HandlerBuilder].
	Default FlagValues
}

// Validate determines whether a [FlagDef] is valid or not.
func (def FlagDef) Validate(_ context.Context, path []string) error {
	var errs error
	isShortFlag := flagKeyIsShortFlag(def.Name)
	shortFlagBecauseOfAlias := ""
	errs = errors.Join(errs, validateFlagKey(def.Name, path))
	for _, alias := range def.Aliases {
		if !isShortFlag && flagKeyIsShortFlag(alias) {
			isShortFlag = true
			shortFlagBecauseOfAlias = alias
		}
		errs = errors.Join(errs, validateFlagKey(alias, path))
	}
	if isShortFlag && !def.IsToggle {
		errs = errors.Join(errs, ShortFlagIsNotToggleError{
			FlagKey:        def.Name,
			Path:           path,
			BecauseOfAlias: shortFlagBecauseOfAlias,
		})
	}
	if def.OnlyToggle && !def.IsToggle {
		errs = errors.Join(errs, OnlyToggleWithoutIsToggleError{
			FlagKey: def.Name,
			Path:    path,
		})
	}
	return errs
}

var (
	flagKeyRE           = regexp.MustCompile(`^--[a-zA-Z0-9][a-zA-Z0-9-_:]*$`)
	shortFlagKeyRE      = regexp.MustCompile(`^-[a-zA-Z0-9]$`)
	symbolOnlyFlagKeyRE = regexp.MustCompile(`^--?[-_:]*$`)
)

func flagKeyIsShortFlag(in string) bool {
	return strings.HasPrefix(in, "-") && !strings.HasPrefix(in, "--")
}

func validateFlagKey(key string, path []string) error {
	if flagKeyRE.MatchString(key) {
		return nil
	}
	if shortFlagKeyRE.MatchString(key) {
		return nil
	}
	if key[0] != '-' {
		return InvalidFlagKeyError{
			CommandPath: path,
			FlagKey:     key,
		}
	}
	if len(key) > 2 && key[1] != '-' {
		return InvalidFlagKeyError{
			CommandPath: path,
			FlagKey:     key,
		}
	}
	if symbolOnlyFlagKeyRE.MatchString(key) {
		return InvalidFlagKeyError{
			CommandPath: path,
			FlagKey:     key,
		}
	}
	return InvalidFlagKeyError{
		CommandPath: path,
		FlagKey:     key,
	}
}
