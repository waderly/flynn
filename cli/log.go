package main

import (
	"io"
	"os"
	"strconv"

	"github.com/flynn/flynn/Godeps/_workspace/src/github.com/flynn/go-docopt"
	"github.com/flynn/flynn/controller/client"
	"github.com/flynn/flynn/pkg/cluster"
)

func init() {
	register("log", runLog, `
usage: flynn log [options] <job>

Stream log for a specific job.

Options:
    -s, --split-stderr  send stderr lines to stderr
    -f, --follow        stream new lines after printing log buffer
    -n <num>            limit log buffer to a certain number of lines
`)
}

func runLog(args *docopt.Args, client *controller.Client) error {
	num, err := strconv.Atoi(args.String["-n"])
	if err != nil {
		num = 0
	}
	rc, err := client.GetJobLog(mustApp(), args.String["<job>"], args.Bool["--follow"], num)
	if err != nil {
		return err
	}
	var stderr io.Writer = os.Stdout
	if args.Bool["--split-stderr"] {
		stderr = os.Stderr
	}
	attachClient := cluster.NewAttachClient(struct {
		io.Writer
		io.ReadCloser
	}{nil, rc})
	attachClient.Receive(os.Stdout, stderr)
	return nil
}
