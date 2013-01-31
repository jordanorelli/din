package din

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// a Command defines an entry point for command-line executables, to be used as
// subcommands to the application's main command.  E.g., a Din project named
// "stanley" might have a command named "runserver", which would be invoked on
// the command line as "stanley runserver"
type Command struct {
	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// short description shown when listing all commands
	Short string

	// long description shown when getting documentation about the command
	Long string

	// Function to execute arguments passed to the command.  Two io.Writers are
	// also provided, representing log and error output destinations.
	Run func(*Command, []string)

	// contains options specific to this subcommand
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own flag parsing.
	CustomFlags bool
}

// returns the name of the subcommand.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Usage prints out the command's usage line to stderr and aborts.
func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	os.Exit(2)
}

// Runnable reports whether the command can be run; otherwise it is a
// documentation pseudo-command
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Bail is a convenience function for aborting a command.  The error is written
// to stderr and the application exists.
func (c *Command) Bail(err error) {
	s := strings.TrimRight(err.Error(), " \n") + "\n"
	os.Stderr.WriteString(s)
	os.Exit(2)
}

// commandSet is used to group commands to be used by a user.  The commands are
// sorted lexicographically upon insertion.
type commandSet struct {
	items   []Command
	nameMax int
}

// satifying sort.Interface to make the commandSet sortable. -------------------
func (c commandSet) Len() int           { return len(c.items) }
func (c commandSet) Less(i, j int) bool { return c.items[i].Name() < c.items[j].Name() }
func (c commandSet) Swap(i, j int)      { c.items[j], c.items[i] = c.items[i], c.items[j] }

// done satisfying sort.Interface ----------------------------------------------

func (c commandSet) getCommand(name string) *Command {
	for _, cmd := range c.items {
		if cmd.Name() == name {
			return &cmd
		}
	}
	return nil
}

// run runs the arguments presented to the commandSet.  This is typically going
// to be the main entrypoint into a din application, since startserver should
// be implmented as a subcommand, such that a developer can override the
// default functionality should they so chose.
func (c commandSet) run(args []string) {
	cmd := c.getCommand(args[0])
	if cmd == nil {
		panic("no command found")
	}
	if cmd.Runnable() {
		if cmd.CustomFlags {
			args = args[1:]
		} else {
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
		}
		cmd.Run(cmd, args)
		return
	}
	fmt.Println(strings.Trim(cmd.Long, " \n"))
}

// holds all of the known commands, to be populated at runtime.
var cmdRegistry commandSet

// RegisterCommand registers a command for usage by the framework.  This is
// typically called inside of init() to establish the command set before the
// framework parses any command-line arguments.
func RegisterCommand(cmd Command) {
	cmdRegistry.items = append(cmdRegistry.items, cmd)
	sort.Sort(cmdRegistry)
	l := len(cmd.Name())
	if l > cmdRegistry.nameMax {
		cmdRegistry.nameMax = l
	}
}

func init() {
	// something of a meta-command; list-commands will list all commands available
	// to din and then exist.
	RegisterCommand(Command{
		UsageLine: "list-commands",
		Short:     "lists all available commands",
		Long: `
the list-commands subcommand lists all available commands that have been registered with the Din framework core, sorted alphabetically by their subcommand name.  Also included is the command's short description.
`,
		Run: func(cmd *Command, args []string) {
			t := "%-" + strconv.Itoa(cmdRegistry.nameMax) + "s : %s\n"
			for _, cmd := range cmdRegistry.items {
				fmt.Printf(t, cmd.Name(), strings.Trim(cmd.Short, " \n"))
			}
		},
	})
}
