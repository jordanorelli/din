package din

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

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

func (c *Command) Bail(err error) {
	fmt.Println("fuck we're bailing")
	fmt.Println(err)
	s := strings.TrimRight(err.Error(), " \n") + "\n"
	os.Stderr.WriteString(s)
	os.Exit(2)
}

type commandSet struct {
	items   []Command
	nameMax int
}

// satifying sort.Interface to make the commandSet sortable. -------------------
func (c commandSet) Len() int           { return len(c.items) }
func (c commandSet) Less(i, j int) bool { return c.items[i].Name() < c.items[j].Name() }
func (c commandSet) Swap(i, j int)      { c.items[j], c.items[i] = c.items[i], c.items[j] }

// done satisfying sort.Interface ----------------------------------------------

// run runs the arguments presented to the commandSet.  This is typically going
// to be the main entrypoint into a din application, since startserver should
// be implmented as a subcommand, such that a developer can override the
// default functionality should they so chose.
func (c commandSet) run(args []string) {
	for _, cmd := range c.items {
		if cmd.Name() == args[0] {
			if cmd.Runnable() {
				cmd.Run(&cmd, args[1:])
				return
			} else {
				fmt.Println(strings.Trim(cmd.Long, " \n"))
				return
			}
		}
	}
	panic("no command found")
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
	// something of a meta-command; listcommands will list all commands available
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
