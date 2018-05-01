// Copyright (c) 2018 codeliveroil. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package niceflags

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	// Create flag set
	flags := NewFlags(
		"pping",
		"pping - Protocol Ping",
		"Tool to simulate TCP and UDP pings. This can also be used as a port scanner.",
		"[options] host port", // Note that you don't specify the command name here
		"help",
		false)

	// Optionally add some examples
	// Note that you don't specify the command name here
	flags.Examples = []string{
		"-s 128 google.com 80",
		"-p udp -c 5 -t 1000 myserver.com 8085",
	}

	// Describe flags with the 'flags' object as you would with
	// the standard flag package
	// Note the usage of the special back-quoted `default` formatter along with the
	// standard back-quoted parameter types.
	flags.Int("s", 64, "Payload `size` in bytes `default`.")
	flags.Int("i", 1000, "Interval `time` between pings in ms `default`.")
	flags.Int("t", 10000, "Max `time`-to-live for each ping (in ms) before moving on to "+
		"the next attempt `default`.")
	flags.String("p", "tcp", "Specify `protocol` to use. Valid values are `default`:\n"+
		"- tcp: also supports 4 or 6 only counterparts.\n"+
		"- udp: also supports 4 or 6 only counterparts.")
	wait := flags.Bool("w", false, "Wait for a response from the server. Ideally, this should be "+
		"set when the protocol is set to udp.")
	max := flags.Int("c", math.MaxInt32, "Stop after sending specified `num`ber of pings.")
	dns := flags.String("d", "", "DNS `server` IP address to use. This can be specified for name "+
		"resolution on systems that don't use the traditional DNS server configurations such as "+
		"/etc/resolv.conf.")

	exp := "pping - Protocol Ping\n" +
		"  Tool to simulate TCP and UDP pings. This can also be used as a port\n" +
		"  scanner.\n" +
		"\n" +
		"Usage: pping [options] host port\n" +
		"\n" +
		"Options:\n" +
		"  -c num       Stop after sending specified number of pings.\n" +
		"  -d server    DNS server IP address to use. This can be specified for\n" +
		"               name resolution on systems that don't use the traditional\n" +
		"               DNS server configurations such as /etc/resolv.conf.\n" +
		"  -i time      Interval time between pings in ms (default=1000).\n" +
		"  -p protocol  Specify protocol to use. Valid values are (default=tcp):\n" +
		"               - tcp: also supports 4 or 6 only counterparts.\n" +
		"               - udp: also supports 4 or 6 only counterparts.\n" +
		"  -s size      Payload size in bytes (default=64).\n" +
		"  -t time      Max time-to-live for each ping (in ms) before moving on\n" +
		"               to the next attempt (default=10000).\n" +
		"  -w           Wait for a response from the server. Ideally, this should\n" +
		"               be set when the protocol is set to udp.\n" +
		"\n" +
		"Examples:\n" +
		"  pping -s 128 google.com 80\n" +
		"  pping -p udp -c 5 -t 1000 myserver.com 8085\n"

	// Test format
	if got := flags.HelpText(); exp != got {
		minLen := int(math.Min(float64(len(exp)), float64(len(got))))
		diffIndex := 0
		expDiff := ""
		gotDiff := ""
		msg := ""
		for diffIndex = 0; diffIndex < minLen; diffIndex++ {
			cExp := exp[diffIndex]
			cGot := got[diffIndex]
			expDiff += fmt.Sprintf("%c", cExp)
			gotDiff += fmt.Sprintf("%c", cGot)
			if cExp != cGot {
				msg = fmt.Sprintf(" (%q vs %q)", cExp, cGot)
				break
			}
		}

		t.Errorf("help text doesn't match at position %d%s. Length: exp=%d, got=%d:\n%s--vs--\n%s", diffIndex, msg, len(exp), len(got), expDiff, gotDiff)

	}

	// Test drop-in functionality
	if err := flags.Parse(strings.Split("-c 35 -d 8.8.8.8 -w", " ")); err != nil {
		t.Fatal("error when parsing", err)
	}

	compare(t, 35, *max)
	compare(t, "8.8.8.8", *dns)
	compare(t, true, *wait)
	compare(t, false, flags.AskingHelp())
}

func TestHelpChecker(t *testing.T) {
	flags := NewFlags("", "", "", "", "helpme", false)
	flags.Parse([]string{"-helpme"})
	compare(t, true, flags.AskingHelp())
}

func compare(t *testing.T, exp, got interface{}) {
	if exp != got {
		t.Errorf("expected: %v, got: %v", exp, got)
	}
}
