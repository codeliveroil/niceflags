// Copyright (c) 2018 codeliveroil. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package niceflags

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
)

const maxLineLength = 72

// Flags defines the standard Flagset from
// the flag package as well as some custom
// fields.
type Flags struct {
	*flag.FlagSet

	// Title is the title of the application.
	Title string

	// Description describes the application.
	Description string

	// UsageOptions indicates how the application must be used
	// in conjunction with the options.
	// Don't specify the command name in the usage options. For example,
	// for a port scanner, you would just specify "[options] host port";
	// so when the usage is printed, it'll be printed as:
	// portscanner [options] host port
	// You can optionally include a usage description succeeding the first line,
	// separated by new lines and this description will be wrapped to the
	// terminal length.
	UsageOptions string

	// Examples defines examples of the usage. Just as in UsageOptions,
	// do not specify the command name in the examples.
	Examples []string

	// PrintAllDefaults prints the default value for a flag if the default
	// value is not the Zero value. If this is set, it will override the
	// back-quoted `default` option that may be embedded in the flag's
	// usage.
	PrintAllDefaults bool

	helpFlagName string
	cmdName      string
}

// NewFlags constructs a new flag-set which can render cleaner help screen
// formatting as opposed to that offered by the standard flag package.
// The parameters are explained in the documentation for the Flags struct.
//
// title, description and usageOptions may be omitted with an emptry string
// but a helpFlagName (e.g. "help") is mandatory. Then, invoking -help will
// display the help screen (on invocation of Help()).
// Don't specify the command name in the usage options. For example,
// for a port-scanner app, you would just specify "[options] host port".
//
// Describe individual flags with the returned Flags object as you would with
// the standard flag package.
// Notes:
//   - Don't describe the help flag. This will be created automatically
//     using helpFlagName.
//   - Type names can be specified within double back-quotes in the flag
//     usage/description. These will be displayed right next to the flag
//     name.
//   - The back-quoted literal `default` can be placed anywhere in the flag
//     description/usage and niceflags will replace that with the non-Zero
//     default value for the flag.
func NewFlags(cmdName, title, description, usageOptions, helpFlagName string, printAllDefaults bool) *Flags {
	cmdName = path.Base(cmdName)
	flags := &Flags{
		flag.NewFlagSet(cmdName, flag.ExitOnError),
		title, description, usageOptions, nil, printAllDefaults,
		helpFlagName, cmdName,
	}
	flags.Bool(flags.helpFlagName, false, "Help screen.")

	flags.Usage = func() {
		PrintErr("See '%s -%s'\n", cmdName, helpFlagName)
	}
	return flags
}

// AskingHelp returns true if the help flag
// has been invoked
func (f *Flags) AskingHelp() bool {
	fl := f.Lookup(f.helpFlagName)
	if fl == nil {
		PrintErr("Help invoked but the option is not implemented.")
		return false
	}
	return fl.Value.String() == "true"

}

// Help prints the help screen and exits if the help flag has
// been invoked.
func (f *Flags) Help() {
	if f.AskingHelp() {
		f.PrintHelp()
		os.Exit(0)
	}
}

// PrintHelp prints the help screen.
func (f *Flags) PrintHelp() {
	PrintErr(sanitize(f.HelpText()))
}

// HelpText returns the help text.
// % signs in any user given text is prefixed with another % (i.e. %%) so
// that they are escaped if passed to a formatter like Printf or Sprintf.
func (f *Flags) HelpText() string {
	var buf bytes.Buffer

	write := func(msg string, args ...interface{}) {
		buf.WriteString(fmt.Sprintf(msg, args...))
	}

	// Title
	if f.Title != "" {
		write(sanitize(f.Title) + "\n")
	}

	pad := func(s string, l int) string {
		s2 := s
		for i := 0; i < (l - len(s)); i++ {
			s2 += " "
		}
		return s2
	}

	wrapText := func(desc string, indentLen int, lineLen int, indentFirstLine bool) {
		firstLine := true
		indent := pad("", indentLen)
		for i, line := range strings.Split(desc, "\n") {
			firstWord := true
			writeLn := func(ln string) {
				firstLine = false
				firstWord = true
				write(ln + "\n")
			}
			var ln string
			if indentFirstLine || i > 0 {
				ln = indent
			}
			tokens := strings.Split(line, " ")
			for _, word := range tokens {
				length := len(ln)
				if firstLine && !indentFirstLine {
					length += len(indent)
				}

				if length+len(word)+1 > maxLineLength {
					writeLn(ln)
					ln = indent
				}
				if !firstWord {
					ln += " "
				}
				firstWord = false
				ln += word
			}
			writeLn(ln)
		}
	}

	// Description
	if f.Description != "" {
		wrapText(sanitize(f.Description), 2, maxLineLength, true)
		write("\n")
	}

	// Command usage
	usageTokens := strings.Split(sanitize(f.UsageOptions), "\\n")
	write("Usage: %s %s\n", f.cmdName, usageTokens[0])
	if l := len(usageTokens); l > 1 {
		rem := strings.Join(usageTokens[1:l], "\n")
		wrapText(rem, 2, maxLineLength, true)
	}

	// Option/Flag details
	write("\nOptions:\n")
	maxFlagLen := 0
	maxParamLen := 0
	var flags [][3]string

	computeFormat := func(fl *flag.Flag) {
		if fl.Name == f.helpFlagName {
			// skip the help command because it may not be a single character command and
			// it'll unnecessarily clutter the help screen.
			return
		}

		if l := len(fl.Name); l > maxFlagLen {
			maxFlagLen = l
		}

		param := ""
		usage := sanitize(fl.Usage)
		if !isZeroValue(fl, fl.DefValue) {
			if f.PrintAllDefaults {
				usage = strings.Replace(usage, "`default`", "", -1)
				usage += fmt.Sprintf("\n[default=%v]", fl.DefValue)
			} else {
				usage = strings.Replace(usage, "`default`", fmt.Sprintf("(default=%v)", fl.DefValue), -1)
			}

		}

		i1 := strings.Index(usage, "`")
		if i1 != -1 {
			i2 := strings.Index(usage[i1+1:], "`")
			if i2 != -1 {
				param = fl.Usage[i1+1 : i1+i2+1]
				usage = strings.Replace(usage, "`", "", 2)
			}
		}
		if l := len(param); l > maxParamLen {
			maxParamLen = l
		}

		flags = append(flags, [3]string{fl.Name, param, usage})
	}

	f.VisitAll(computeFormat)

	for _, fl := range flags {
		s := fmt.Sprintf("  -%s ", pad(fl[0], maxFlagLen))
		s += fmt.Sprintf("%s  ", pad(fl[1], maxParamLen))
		write(s)
		wrapText(fl[2], len(s), maxLineLength, false)
	}

	// Examples
	if f.Examples != nil && len(f.Examples) > 0 {
		write("\nExamples:\n")
		for _, e := range f.Examples {
			write("  %s %s\n", f.cmdName, sanitize(e))
		}
	}

	return buf.String()
}

// isZeroValue guesses whether the string represents the zero
// value for a flag. It is not accurate but in practice works OK.
// This is a direct copy from the flag package
func isZeroValue(fl *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(fl.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	if value == z.Interface().(flag.Value).String() {
		return true
	}

	switch value {
	case "false", "", "0":
		return true
	}
	return false
}

func sanitize(msg string) string {
	return strings.Replace(msg, "%", "%%", -1)
}

// PrintErr prints to stderr because that's where
// 'flags.PrintAllDefaults()' prints to
func PrintErr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...) //stderr because
}
