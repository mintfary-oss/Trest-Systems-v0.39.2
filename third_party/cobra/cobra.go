// Package cobra provides the small command/flag subset used by Trest Systems.
// It is an internal compatibility implementation built on the Go standard
// library so release binaries can be produced without downloading modules.
package cobra

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Command struct {
	Use          string
	Short        string
	Long         string
	SilenceUsage bool
	RunE         func(cmd *Command, args []string) error

	parent     *Command
	children   []*Command
	persistent *FlagSet
	local      *FlagSet
	args       []string
	ctx        context.Context
	out        io.Writer
	err        io.Writer
}

type flagKind int

const (
	kindString flagKind = iota
	kindDuration
	kindBool
)

type flagDef struct {
	name     string
	kind     flagKind
	stringP  *string
	durP     *time.Duration
	boolP    *bool
	usage    string
	defValue string
}

type FlagSet struct {
	defs map[string]*flagDef
}

func newFlagSet() *FlagSet { return &FlagSet{defs: make(map[string]*flagDef)} }

func (f *FlagSet) StringVar(p *string, name, value, usage string) {
	*p = value
	f.defs[name] = &flagDef{name: name, kind: kindString, stringP: p, usage: usage, defValue: value}
}

func (f *FlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	*p = value
	f.defs[name] = &flagDef{name: name, kind: kindDuration, durP: p, usage: usage, defValue: value.String()}
}

func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	*p = value
	f.defs[name] = &flagDef{name: name, kind: kindBool, boolP: p, usage: usage, defValue: fmt.Sprint(value)}
}

func (c *Command) PersistentFlags() *FlagSet {
	if c.persistent == nil {
		c.persistent = newFlagSet()
	}
	return c.persistent
}

func (c *Command) Flags() *FlagSet {
	if c.local == nil {
		c.local = newFlagSet()
	}
	return c.local
}

func (c *Command) AddCommand(children ...*Command) {
	for _, child := range children {
		if child == nil {
			continue
		}
		child.parent = c
		c.children = append(c.children, child)
	}
}

func (c *Command) SetArgs(args []string)          { c.args = append([]string(nil), args...) }
func (c *Command) SetOut(w io.Writer)             { c.out = w }
func (c *Command) SetErr(w io.Writer)             { c.err = w }
func (c *Command) SetContext(ctx context.Context) { c.ctx = ctx }

func (c *Command) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	if c.parent != nil {
		return c.parent.Context()
	}
	return context.Background()
}

func (c *Command) OutOrStdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	if c.parent != nil {
		return c.parent.OutOrStdout()
	}
	return os.Stdout
}

func (c *Command) ErrOrStderr() io.Writer {
	if c.err != nil {
		return c.err
	}
	if c.parent != nil {
		return c.parent.ErrOrStderr()
	}
	return os.Stderr
}

func (c *Command) Execute() error {
	args := c.args
	if args == nil {
		args = os.Args[1:]
	}
	return c.execute(args)
}

func (c *Command) execute(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		c.printUsage()
		return nil
	}
	if c.parent == nil {
		cleaned, err := applyFlags(args, c.PersistentFlags(), true)
		if err != nil {
			return err
		}
		args = cleaned
	}

	if len(args) > 0 {
		for _, child := range c.children {
			name := commandName(child.Use)
			if args[0] == name {
				child.ctx = c.Context()
				child.out = c.out
				child.err = c.err
				return child.executeChild(args[1:])
			}
		}
	}

	if c.RunE != nil {
		return c.RunE(c, args)
	}
	if len(c.children) > 0 {
		c.printUsage()
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q", args[0])
		}
		return errors.New("a command is required")
	}
	return nil
}

func (c *Command) executeChild(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		c.printUsage()
		return nil
	}
	// Cobra accepts persistent flags on either side of the subcommand.
	root := c
	for root.parent != nil {
		root = root.parent
	}
	var err error
	args, err = applyFlags(args, root.PersistentFlags(), true)
	if err != nil {
		return err
	}
	args, err = applyFlags(args, c.Flags(), false)
	if err != nil {
		return err
	}
	if c.RunE == nil {
		return nil
	}
	return c.RunE(c, args)
}

func applyFlags(args []string, defs *FlagSet, ignoreUnknown bool) ([]string, error) {
	if defs == nil || len(defs.defs) == 0 {
		return append([]string(nil), args...), nil
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			out = append(out, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "--") || a == "--" {
			out = append(out, a)
			continue
		}
		nameValue := strings.TrimPrefix(a, "--")
		name, value, hasValue := strings.Cut(nameValue, "=")
		def, ok := defs.defs[name]
		if !ok {
			if ignoreUnknown {
				out = append(out, a)
				continue
			}
			return nil, fmt.Errorf("unknown flag --%s", name)
		}
		if def.kind == kindBool && !hasValue {
			value, hasValue = "true", true
		}
		if !hasValue {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag --%s requires a value", name)
			}
			i++
			value = args[i]
		}
		if err := def.set(value); err != nil {
			return nil, fmt.Errorf("invalid value for --%s: %w", name, err)
		}
	}
	return out, nil
}

func (d *flagDef) set(value string) error {
	switch d.kind {
	case kindString:
		*d.stringP = value
		return nil
	case kindDuration:
		v, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d.durP = v
		return nil
	case kindBool:
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		*d.boolP = v
		return nil
	default:
		return flag.ErrHelp
	}
}

func parseBool(v string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "0", "f", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected boolean")
	}
}

func commandName(use string) string {
	if fields := strings.Fields(use); len(fields) > 0 {
		return fields[0]
	}
	return use
}

func (c *Command) printUsage() {
	out := c.ErrOrStderr()
	fmt.Fprintf(out, "Usage: %s <command> [flags]\n", commandName(c.Use))
	if c.Short != "" {
		fmt.Fprintln(out, c.Short)
	}
	if len(c.children) > 0 {
		fmt.Fprintln(out, "Commands:")
		for _, child := range c.children {
			fmt.Fprintf(out, "  %-12s %s\n", commandName(child.Use), child.Short)
		}
	}
}
