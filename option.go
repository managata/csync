//
//
//

package main

import (
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	SrcDir string `short:"s" long:"source-dir" description:"source directory"`
	DstDir string `short:"d" long:"destination-dir" description:"destination directory"`

	Route    []string `short:"r" long:"route" description:"src_host:dst_host" default:"localhost:localhost"`
	Parallel int      `short:"p" long:"parallel" default:"1" description:"number of rsync processes per route"`

	//
	MkdirCommand string   `short:"m" long:"mkdir-command" description:"mkdir command" default:"mkdir"`
	MkdirOptions []string `short:"n" long:"mkdir-option" description:"mkdir options" default:"-p"`

	SshCommand string   `short:"a" long:"ssh-command" description:"ssh command" default:"ssh"`
	SshOptions []string `short:"b" long:"ssh-option" description:"ssh options" default:"-tt"`
	//	SshOption  []string `short:"b" long:"ssh-option" description:"ssh options" default:"{\"-l\", \"mana\"}"`

	RsyncCommand string   `short:"x" long:"rsync-command" description:"rsync command" default:"rsync"`
	RsyncOptions []string `short:"y" long:"rsync-option" description:"rsync options" default:"-dglopstADHX" default:"--delete"`

	//

	LogFile  string `short:"o" long:"log-file" description:"write logs to file"`
	LogLevel int    `short:"O" long:"log-level" description:"[-1|0|1|2|3|4|5|6|7] log level" default:"5"`
	Verbose  bool   `short:"v" long:"verbose" description:"same as --log-level=6"`
	Quiet    bool   `short:"q" long:"quiet" description:"same as --log-level=-1"`

	//

	//	IgnoreError bool `long:"ignore-error" description:"continue as much as possible"`
	ExitAtWarn bool `short:"w" long:"exit-at-warn" description:"treat continuable warning as error"`

	//

	Version bool `long:"version" description:"print version"`

	//

	StdIn  bool `no-flag:"true" default:"false"`
	StdOut bool `no-flag:"true" default:"false"`
}

var oP Options
var version string

//

func parseFlags() {
	parser := flags.NewParser(&oP, flags.Default)
	parser.Name = "csync"
	parser.Usage = `[OPTIONS...] -s src_dir -d dst_dir`
	_, err := parser.Parse()

	if err != nil {
		//		fmt.Fprint(os.Stderr, "csync: %s\n", err)
		os.Exit(1)
	}

	if oP.Version {
		fmt.Fprintf(os.Stderr, "csync: %s\n", version)
		os.Exit(0)
	}

	//
	if oP.Verbose {
		oP.LogLevel = 6
	}
	if oP.Quiet {
		oP.LogLevel = -1
	}

	if len(oP.SrcDir) == 0 {
		fmt.Fprintf(os.Stderr, "csync: source directory not sprcified\n")
		os.Exit(1)
	}
	if len(oP.DstDir) == 0 {
		fmt.Fprintf(os.Stderr, "csync: destination directory not sprcified\n")
		os.Exit(1)
	}

	for _, v := range oP.Route {
		if !strings.Contains(v, ":") {
			fmt.Fprintf(os.Stderr, "csync: '%s' does not contain separater mark ':'\n", v)
			os.Exit(1)
		}
	}

}
