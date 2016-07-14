package sv

import (
	"errors"
	// "github.com/codegangsta/cli"
	// "github.com/codegangsta/envy/lib"
	"github.com/codegangsta/gin/lib"
	"github.com/fvbock/endless"
	"github.com/kardianos/osext"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	// "strconv"
	"syscall"
	"time"
)

var GraceOptions struct {
	AutoRestart bool
	AutoBuild   bool
}

var restartSignal = syscall.SIGHUP

func GraceListenAndServe(addr string, handler http.Handler) error {
	err := graceConfig();
	if err!=nil {
		return err
	}
	restartSignal = syscall.SIGHUP
	return endless.ListenAndServe(addr, handler)
}

func graceConfig() error{

	if GraceOptions.AutoRestart {
		filename, _ := osext.Executable()
		info, err := os.Stat(filename)
		if err != nil {
			return err
		}
		mtime := info.ModTime()

		go graceAutoRestartCheckRoutin(filename, mtime, restartSignal)
	}

	if GraceOptions.AutoBuild {
		filename, _ := osext.Executable()
		go autoBuildRoutine(".",filename)
	}

	return nil
}

func graceAutoRestartCheckRoutin(filename string, mtime time.Time, sig syscall.Signal) {
	for {
		time.Sleep(time.Second)

		info, err := os.Stat(filename)
		if err != nil {
			continue
		}

		if info.ModTime().After(mtime) {
			Log.Info("Detect mtine change, send restart signal")
			process, _ := os.FindProcess(os.Getpid())
			process.Signal(sig)
			break
		}
	}
}

func autoBuildRoutine(dir string,bin string) {

	wd, err := os.Getwd()
	if err != nil {
		Log.Fatal(err)
	}

	builder := gin.NewBuilder(dir, bin, false)
	runner := gin.NewRunner(filepath.Join(wd, builder.Binary()))
	runner.SetWriter(os.Stdout)
	onShutdown(runner)

	// scan for changes
	scanChanges(dir, func(path string) {
		runner.Kill()
		build(builder, runner)
	})
}

var startTime  = time.Now()
var buildError error

func build(builder gin.Builder, runner gin.Runner) {
	err := builder.Build()
	if err != nil {
		buildError = err
		Log.Error("ERROR! Build failed:\n" + builder.Errors())
	} else {
		// print success only if there were errors before
		if buildError != nil {
			Log.Error("Build Successful")
		}
		buildError = nil
	}

	time.Sleep(100 * time.Millisecond)
}

type scanCallback func(path string)


func scanChanges(watchPath string, cb scanCallback) {
	for {
		filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if path == ".git" {
				return filepath.SkipDir
			}

			// ignore hidden files
			if filepath.Base(path)[0] == '.' {
				return nil
			}

			if filepath.Ext(path) == ".go" && info.ModTime().After(startTime) {
				cb(path)
				startTime = time.Now()
				return errors.New("done")
			}

			return nil
		})
		time.Sleep(500 * time.Millisecond)
	}
}


func onShutdown(runner gin.Runner) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		Log.Info("Got signal: ", s)
		err := runner.Kill()
		if err != nil {
			Log.Error("Error killing: ", err)
		}
		// os.Exit(1)
	}()
}

