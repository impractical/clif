package clif

import (
	"context"
	"fmt"
)

type funcCommandHandler func(ctx context.Context, resp *Response)

func (f funcCommandHandler) Build(_ context.Context, _ []Flag, _ []string, _ *Response) Handler { //nolint:ireturn // filling an interface
	return f
}

func (f funcCommandHandler) Handle(ctx context.Context, resp *Response) {
	f(ctx, resp)
}

type flagCommandHandler struct {
	flags []Flag
	args  []string
	f     func(ctx context.Context, flags []Flag, args []string, resp *Response)
}

func (f flagCommandHandler) Build(_ context.Context, flags []Flag, args []string, _ *Response) Handler { //nolint:ireturn // filling an interface
	f.flags = flags
	f.args = args
	return f
}

func (f flagCommandHandler) Handle(ctx context.Context, resp *Response) {
	f.f(ctx, f.flags, f.args, resp)
}

//nolint:errcheck // several places we're not checking errors because they can't fail
func ExampleApplication() {
	app := Application{
		Commands: []Command{
			{
				Name:        "help",
				Description: "Displays help information about this program.",
				Handler: funcCommandHandler(func(_ context.Context, resp *Response) {
					resp.Output.Write([]byte("this is help information\n"))
				}),
			},
			{
				Name: "foo",
				Subcommands: []Command{
					{
						Name: "bar",
						Flags: []FlagDef{
							{
								Name:                 "quux",
								ValueAccepted:        true,
								OnlyAfterCommandName: false,
								Parser:               StringParser{},
							},
						},
						Handler: flagCommandHandler{
							f: func(_ context.Context, flags []Flag, args []string, resp *Response) {
								fmt.Fprintln(resp.Output, flags, args)
							},
						},
					},
				},
			},
		},
		Flags: []FlagDef{
			{
				Name:   "baaz",
				Parser: BoolParser{},
			},
		},
	}
	res := app.Run(context.Background(), WithArgs([]string{"help"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"foo", "bar", "--quux", "hello"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"foo", "--quux", "hello", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"--quux", "hello", "foo", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"foo", "bar", "--quux=hello"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"foo", "--quux=hello", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"--quux=hello", "foo", "bar"}))
	fmt.Println(res)
	res = app.Run(context.Background(), WithArgs([]string{"--baaz", "foo", "bar", "--quux", "hello"}))
	fmt.Println(res)
	// output:
	// this is help information
	// 0
	// [{quux hello hello}] []
	// 0
	// [{quux hello hello}] []
	// 0
	// [{quux hello hello}] []
	// 0
	// [{quux hello hello}] []
	// 0
	// [{quux hello hello}] []
	// 0
	// [{quux hello hello}] []
	// 0
	// [{baaz  true} {quux hello hello}] []
	// 0
}
