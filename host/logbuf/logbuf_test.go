package logbuf

import (
	"io"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/flynn/flynn/Godeps/_workspace/src/github.com/flynn/lumberjack"
	. "github.com/flynn/flynn/Godeps/_workspace/src/gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestLogWriteRead(c *C) {
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()
	defer stdoutW.Close()
	defer stderrW.Close()

	l := NewLog(&lumberjack.Logger{Dir: c.MkDir()})
	defer l.Close()
	ch := make(chan Data)
	err := l.Read(0, ch)
	c.Assert(err, IsNil)
	c.Assert(len(ch), Equals, 0)

	follow := func(stream int, r io.Reader) {
		if err := l.Follow(stream, r); err != nil && err != io.EOF {
			c.Error(err)
		}
	}
	go follow(1, stdoutR)
	go follow(2, stderrR)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		stdoutW.Write([]byte("1"))
		stdoutW.Write([]byte("2"))
		wg.Done()
	}()
	go func() {
		stderrW.Write([]byte("3"))
		stderrW.Write([]byte("4"))
		wg.Done()
	}()
	wg.Wait()
	ch = make(chan Data)
	go l.Read(0, ch)
	c.Assert(err, IsNil)

	stdout, stderr := 0, 2
	for i := 0; i < 4; i++ {
		var line Data
		select {
		case line = <-ch:
		case <-time.After(time.Second):
			c.Error("timed out")
		}
		c.Assert(line.Timestamp.After(time.Now().Add(-time.Minute)), Equals, true)
		switch line.Stream {
		case 1:
			stdout++
			c.Assert(line.Message, Equals, strconv.Itoa(stdout))
		case 2:
			stderr++
			c.Assert(line.Message, Equals, strconv.Itoa(stderr))
		default:
			c.Errorf("unknown stream: %#v", line)
		}
	}
	/*
		err = l.l.Rotate()
		c.Assert(err, IsNil)

		stdoutW.Write([]byte("5"))
		line, err := r.ReadData(false)
		c.Assert(err, IsNil)
		c.Assert(line.Message, Equals, "5")

		_, err = r.ReadData(false)
		c.Assert(err, Equals, io.EOF)*/
}

func (s *S) TestStreaming(c *C) {
	l := NewLog(&lumberjack.Logger{Dir: c.MkDir()})
	pipeR, pipeW := io.Pipe()
	go l.Follow(1, pipeR)

	ch := make(chan Data)
	l.AddListener(-1, ch)
	defer l.RemoveListener(-1, ch)

	for i := 0; i < 3; i++ {
		s := strconv.Itoa(i)
		pipeW.Write([]byte(s))
		var data Data
		select {
		case data = <-ch:
		case <-time.After(time.Second):
			c.Error("timed out")
		}
		c.Assert(data, Not(IsNil))
		c.Assert(data.Message, Equals, s)
	}

	runtime.Gosched()
	pipeW.Close()
	l.Close()
	<-ch
}
