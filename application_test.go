package clif_test

import (
	"context"
	"fmt"

	"impractical.co/clif"
	"impractical.co/clif/flagtypes"
)

type funcCommandHandler func(ctx context.Context, resp *clif.Response)

func (f funcCommandHandler) Build(_ context.Context, _ map[string]clif.Flag, _ []string, _ *clif.Response) clif.Handler { //nolint:ireturn // filling an interface
	return f
}

func (f funcCommandHandler) Handle(ctx context.Context, resp *clif.Response) {
	f(ctx, resp)
}

type flagCommandHandler struct {
	flags map[string]clif.Flag
	args  []string
	f     func(ctx context.Context, flags map[string]clif.Flag, args []string, resp *clif.Response)
}

func (f flagCommandHandler) Build(_ context.Context, flags map[string]clif.Flag, args []string, _ *clif.Response) clif.Handler { //nolint:ireturn // filling an interface
	f.flags = flags
	f.args = args
	return f
}

func (f flagCommandHandler) Handle(ctx context.Context, resp *clif.Response) {
	f.f(ctx, f.flags, f.args, resp)
}

//nolint:errcheck // several places we're not checking errors because they can't fail
func ExampleApplication() {
	app := clif.Application{
		Commands: []clif.Command{
			{
				Name:        "help",
				Description: "Displays help information about this program.",
				Handler: funcCommandHandler(func(_ context.Context, resp *clif.Response) {
					resp.Output.Write([]byte("this is help information\n"))
				}),
			},
			{
				Name: "foo",
				Subcommands: []clif.Command{
					{
						Name: "bar",
						Flags: []clif.FlagDef{
							{
								Name:                 "quux",
								ValueAccepted:        true,
								OnlyAfterCommandName: false,
								Parser:               flagtypes.StringParser{},
							},
						},
						Handler: flagCommandHandler{
							f: func(_ context.Context, flags map[string]clif.Flag, args []string, resp *clif.Response) {
								fmt.Fprintln(resp.Output, flags, args)
							},
						},
					},
				},
			},
		},
		Flags: []clif.FlagDef{
			{
				Name:   "baaz",
				Parser: flagtypes.BoolParser{},
			},
		},
	}
	res := app.Run(context.Background(), clif.WithArgs([]string{"help"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"foo", "bar", "--quux", "hello"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"foo", "--quux", "hello", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"--quux", "hello", "foo", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"foo", "bar", "--quux=hello"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"foo", "--quux=hello", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"--quux=hello", "foo", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), clif.WithArgs([]string{"--baaz", "foo", "bar", "--quux", "hello"}))
	fmt.Println(res)
	// output:
	// this is help information
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[quux:{quux hello hello}] []
	// 0
	// map[baaz:{baaz  true} quux:{quux hello hello}] []
	// 0
}
