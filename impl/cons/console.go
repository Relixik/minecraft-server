package cons

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"

	"github.com/Relixik/minecraft-server/apis/base"
	"github.com/Relixik/minecraft-server/apis/logs"
	"github.com/Relixik/minecraft-server/impl/data/system"
)

type Console struct {
	i io.Reader
	o io.Writer

	logger *logs.Logging

	IChannel chan string
	OChannel chan string

	report chan system.Message
	
	// Memory leak prevention
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConsole(report chan system.Message) *Console {
	ctx, cancel := context.WithCancel(context.Background())
	
	console := &Console{
		IChannel: make(chan string),
		OChannel: make(chan string),

		report: report,
		
		ctx:    ctx,
		cancel: cancel,
	}

	console.i = io.MultiReader(os.Stdin)
	console.o = io.MultiWriter(os.Stdout, console.newLogFile("latest.log"))

	console.logger = logs.NewLoggingWith("console", console.o, logs.EveryLevel...)

	return console
}

func (c *Console) Load() {
	// handle i channel
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		scanner := bufio.NewScanner(c.i)

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if scanner.Scan() {
					err := base.Attempt(func() {
						c.IChannel <- scanner.Text()
					})

					if err != nil {
						c.report <- system.Make(system.FAIL, err)
					}
				} else {
					// EOF or error, check context
					select {
					case <-c.ctx.Done():
						return
					default:
						continue
					}
				}
			}
		}
	}()

	// handle o channel
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
				return
			case line := <-c.OChannel:
				c.logger.Info(line)
			}
		}
	}()
}

func (c *Console) Kill() {
	c.logger.Info("Shutting down console...")
	
	// Cancel context to signal goroutines to stop
	c.cancel()
	
	// Wait for goroutines to finish
	c.wg.Wait()
	
	defer func() {
		_ = recover() // ignore panic with closing closed channel
	}()

	// save the log file as YYYY-MM-DD-{index}.log{.gz optionally compressed}

	close(c.IChannel)
	close(c.OChannel)
	
	c.logger.Info("Console shutdown complete")
}

func (c *Console) Name() string {
	return "ConsoleSender"
}

func (c *Console) SendMessage(message ...interface{}) {
	defer func() {
		if err := recover(); err != nil {
			c.report <- system.Make(system.FAIL, err)
		}
	}()

	c.OChannel <- base.ConvertToString(message...)
}

type logFileWriter struct {
	file *os.File
}

func (c *Console) newLogFile(name string) io.Writer {
	file, err := os.Create(name)
	if err != nil {
		c.report <- system.Make(system.FAIL, err)
		return nil
	}

	return &logFileWriter{file: file}
}

func (l *logFileWriter) Write(p []byte) (n int, err error) {

	// this is going to be messy, but this should convert to string, strip colors, and then write to file. Don't @ me.

	return l.file.Write(p)
}
