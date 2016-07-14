package sv

// 标准应用包的目录结构

import (
	"os"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"github.com/kardianos/osext"
)

var baseDir string = ""

func init() {
	appInitDir()
}

func AppWritePidFile(path string) {
	pid := os.Getpid()
	ioutil.WriteFile(path,[]byte(strconv.Itoa(pid)),0755)
}

func appInitDir() {

	if(baseDir!="") {
		return;
	}

	execPath,err := osext.Executable()
	if err!=nil {
		panic(err.Error())
	}
	absPath,absErr := filepath.Abs(execPath)
	if absErr!=nil {
		panic(absErr.Error())
	}

	dir := filepath.Dir(absPath)
	if(filepath.Base(dir)=="bin") {
		baseDir, absErr = filepath.Abs(dir + string(filepath.Separator) + "..")
		if absErr!=nil {
			panic(absErr.Error())
		}
	} else {
		baseDir = dir;
	}

	os.Mkdir(AppLogDir(),0755)
	os.Mkdir(AppDataDir(),0755)
}

func AppBaseDir() string{
	appInitDir();
	return baseDir
}

func AppName() string{
	return filepath.Base(AppBaseDir())
}

func AppLogDir() string{
	return AppBaseDir() + string(filepath.Separator) + "log"
}

func AppDataDir() string{
	return AppBaseDir() + string(filepath.Separator) + "data"
}

func AppBinDir() string{
	return AppBaseDir() + string(filepath.Separator) + "bin"
}

func AppConfDir() string{
	return AppBaseDir() + string(filepath.Separator) + "conf"
}
