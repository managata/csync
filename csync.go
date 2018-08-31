//
//
//

package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/sync/errgroup"
)

var rn = regexp.MustCompile("\r\n|\n\r|\n|\r")

func main() {
	parseFlags()

	eOpen()
	defer eClose()

	if !isDirPresent(oP.SrcDir) {
		eMsg(E_CRIT, nil, "\"%s\": source directory not exist", oP.SrcDir)
	}

	var err error
	oP.SrcDir, err = filepath.Abs(oP.SrcDir)
	if err != nil {
		eMsg(E_CRIT, err, "")
	}
	oP.DstDir, err = filepath.Abs(oP.DstDir)
	if err != nil {
		eMsg(E_CRIT, err, "")
	}
	eMsg(E_DEBUG, nil, "SrcDir:%s DstDir:%s", oP.SrcDir, oP.DstDir)

	eg, ctx := errgroup.WithContext(context.TODO())

	pc := make(chan string, oP.Parallel*len(oP.Route)+1)
	// 	pc := make(chan string, 1)
	//	pc := make(chan string, 2)

	n := 0
	for _, v := range oP.Route {
		sd := strings.Split(v, ":")
		if sd[0] == sd[1] {
			sd[0] = ""
		}
		if sd[0] == "localhost" {
			sd[0] = ""
		}
		if sd[1] == "localhost" {
			sd[1] = ""
		}
		for i := 0; i < oP.Parallel; i++ {
			s := sd[0]
			d := sd[1]
			m := n
			eg.Go(func() error {
				return syncDir(ctx, pc, s, d, m)
			})
			n++
		}
	}

	eg.Go(func() error {
		return traceDir(ctx, pc)
	})

	if err := eg.Wait(); err != nil {
		eMsg(E_CRIT, err, "")
	}

	return
}

func traceDir(ctx context.Context, pc chan string) (err error) {
	eMsg(E_DEBUG, nil, "-: .")
	pc <- "."

	err = walkDir(ctx, pc, "")

	close(pc)
	return err
}

func walkDir(ctx context.Context, pc chan string, rdir string) (err error) {
	adir := filepath.Join(oP.SrcDir, rdir)

	var fl []os.FileInfo
	fl, err = ioutil.ReadDir(adir)
	if err != nil {
		eMsg(E_WARN, err, "")
		if oP.ExitAtWarn {
			return
		} else {
			return nil
		}
	}

	for _, v := range fl {
		if v.IsDir() {
			p := filepath.Join(rdir, v.Name())
			eMsg(E_DEBUG, nil, "-: %s", p)

			dc := make(chan int)

			go func() {
				defer func() {
					_ = recover()
				}()
				pc <- p
				dc <- 1
			}()

			select {
			case <-dc:
			case <-ctx.Done():
				eMsg(E_DEBUG, nil, "-: cancel :%s", p)
				return errors.New("cancel waldDir")
			}

			err = walkDir(ctx, pc, p)
			if err != nil {
				return
			}
		}
	}

	return
}

func syncDir(ctx context.Context, pc chan string, shost string, dhost string, n int) (err error) {
	eMsg(E_DEBUG, nil, "syncDir[%d]: start shost:%s dhost:%s", n, shost, dhost)

	var cmd *exec.Cmd
	margs := make([]string, 0, 16)
	rargs := make([]string, 0, 16)

	// mkdir args
	if len(oP.MkdirCommand) != 0 {
		if len(dhost) != 0 {
			margs = append(margs, oP.SshCommand)
			if len(oP.SshOptions) != 0 {
				margs = append(margs, oP.SshOptions[0:]...)
			}
			margs = append(margs, dhost)
		}
		margs = append(margs, oP.MkdirCommand)
		if len(oP.MkdirOptions) != 0 {
			margs = append(margs, oP.MkdirOptions[0:]...)
		}
	}

	// rsync args
	if len(dhost) != 0 {
		rargs = append(rargs, oP.SshCommand)
		if len(oP.SshOptions) != 0 {
			rargs = append(rargs, oP.SshOptions[0:]...)
		}
		rargs = append(rargs, dhost)
	}
	rargs = append(rargs, oP.RsyncCommand)
	if len(oP.RsyncOptions) != 0 {
		rargs = append(rargs, oP.RsyncOptions[0:]...)
	}

	//

	for {
		rdir, ok := <-pc
		if !ok {
			eMsg(E_DEBUG, nil, "syncDir[%d]: close", n)
			return
		}

		dc := make(chan int)

		go func() {
			sdir := filepath.Join(oP.SrcDir, rdir) + "/"
			ddir := filepath.Join(oP.DstDir, rdir) + "/"
			eMsg(E_DEBUG, nil, "syncDir[%d]: dir:%s sdir:%s ddir:%s", n, rdir, sdir, ddir)
			eMsg(E_INFO, nil, "copying %s", sdir)

			// rsync on destination host
			// * l:l,              mkdir 'dst';       rsync -@   src dst
			//   l:d,        ssh d mkdir 'dst'; ssh d rsync -@   src dst
			// * d:d -> l:d, ssh d mkdir 'dst'; ssh d rsync -@   src dst
			// * s:d,        ssh d mkdir 'dst'; ssh d rsync -@ s:src dst
			// * s:l,              mkdir 'dst';       rsync -@ s:src dst

			// mkdir
			if len(margs) != 0 {
				args := append(margs, ddir)
				eMsg(E_DEBUG, nil, "syncDir[%d]: %v", n, args)

				cmd = exec.Command(args[0], args[1:]...)
				out, err := cmd.CombinedOutput()
				es := 0
				if ee, ok := err.(*exec.ExitError); ok {
					if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
						es = ws.ExitStatus()
					}
				}

				for _, v := range rn.Split(string(out), -1) {
					eMsg(E_DEBUG, nil, "syncDir[%d]: mkdir es:%d out:%s", n, es, v)
					e := E_ERR
					if es == 0 {
						e = E_INFO
					}
					if len(v) != 0 {
						eMsg(e, nil, "%s", v)
					}
				}

				if es != 0 {
					dc <- es
					return
				}
			}

			// rsync
			src := sdir
			if len(shost) != 0 {
				src = fmt.Sprintf("%s:%s", shost, sdir)
			}

			args := append(rargs, src, ddir)
			eMsg(E_DEBUG, nil, "syncDir[%d]: %v", n, args)

			cmd = exec.Command(args[0], args[1:]...)
			out, err := cmd.CombinedOutput()
			es := 0
			if ee, ok := err.(*exec.ExitError); ok {
				if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
					es = ws.ExitStatus()
				}
			}

			for _, v := range rn.Split(string(out), -1) {
				eMsg(E_DEBUG, nil, "syncDir[%d]: rsync es:%d out:%s", n, es, v)
				e := E_ERR
				switch es {
				case 0:
					e = E_INFO
				case 23, 24:
					e = E_WARN
				}
				if len(v) != 0 {
					eMsg(e, nil, "%s", v)
				}
			}

			dc <- es
			return
		}()

		select {
		case es := <-dc:
			eMsg(E_DEBUG, nil, "syncDir[%d]: done es:%d", n, es)
			switch es {
			case 0:
			case 23, 24:
				if oP.ExitAtWarn {
					return errors.New("csync terminated because of warning")
				}
			default:
				return errors.New("csync terminated because of error")
			}

		case <-ctx.Done():
			if (cmd != nil) && (cmd.Process != nil) {
				cmd.Process.Kill()
				eMsg(E_DEBUG, nil, "syncDir[%d]: cancel pid:%d", n, cmd.Process.Pid)
			} else {
				eMsg(E_DEBUG, nil, "syncDir[%d]: cancel", n)
			}
			return
		}

		eMsg(E_DEBUG, nil, "syncDir[%d]: end", n)
	}
}
