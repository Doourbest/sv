package sv

import (
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/Sirupsen/logrus"
	"github.com/rifflock/lfshook"
)

// provide the default logger
var Log *logrus.Logger = nil

func init() {

	logDir := AppLogDir()
	logName := filepath.Base(os.Args[0])

	Log = logrus.New()
	Log.Out = ioutil.Discard; // do not write to stderr
	f := new(logrus.TextFormatter)
	f.ForceColors = false;
	f.DisableColors = true;
	Log.Formatter = f
	Log.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.DebugLevel : logDir + string(filepath.Separator) + logName + ".log",
		logrus.InfoLevel  : logDir + string(filepath.Separator) + logName + ".log",
		logrus.WarnLevel  : logDir + string(filepath.Separator) + logName + ".log",
		logrus.ErrorLevel : logDir + string(filepath.Separator) + logName + ".log",
		logrus.FatalLevel : logDir + string(filepath.Separator) + logName + ".log",
		logrus.PanicLevel : logDir + string(filepath.Separator) + logName + ".log",
	}))

	return
}

