package clif

import (
	"context"
)

// VersionMiddleware provides a middleware that, when any of the specified
// flags are passed, will print the version output and prevent any other
// [Handler] or [Middleware] from executing.
type VersionMiddleware struct {
	Flags []string
}

func (middleware VersionMiddleware) Intercept(ctx context.Context, flags FlagSet, _ []string, defs DefParams, resp *Response) bool { //nolint:revive // that's the way interfaces go sometimes, this really needs 5 arguments
	var showVersion bool
	for _, flag := range middleware.Flags {
		versionFlag, ok := flags[flag]
		if !ok {
			continue
		}
		err := versionFlag.As(ctx, &showVersion)
		if err != nil {
			fprintlnOrPanic(resp.Error, "Error checking if the", versionFlag, "flag is set:", err.Error())
			resp.Code = 1
			return false
		}
		if showVersion {
			break
		}
	}
	if !showVersion {
		return true
	}
	fprintlnOrPanic(resp.Output, defs.App.Name, defs.App.Version)
	return false
}
