package sv

// 标准应用包的目录结构

import (
	"os"
	"fmt"
	"reflect"
	"strings"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"github.com/kardianos/osext"
	"github.com/doourbest/sv/utils"
	// "github.com/urfave/cli"
	"github.com/BurntSushi/toml"
)

var baseDir string = ""

func init() {
	appInitDir()
}

func AppStart(app interface{}) error {

	v := reflect.ValueOf(app)
	if v.Kind()!=reflect.Ptr || v.IsNil() || v.Elem().Kind()!=reflect.Struct {
		return fmt.Errorf("AppStart parameter should be pointer to struct, %s given", v.Type().String())
	}

	// parse command method
	if len(os.Args)<=1 {
		return fmt.Errorf("You should specify a command to execute")
	}
	cmd := os.Args[1]
	mn := "Cmd"+utils.UcFirst(cmd)
	m := v.MethodByName(mn)
	if !m.IsValid() {
		return fmt.Errorf("No such command [%s]", cmd)
	}
	numIn := m.Type().NumIn()
	if numIn>2 {
		return fmt.Errorf("No such command [%s] because [func %s declaration is invalid, too much parameters required]", cmd, mn)
	}
	params := make([]reflect.Value,numIn,numIn)
	for i:=0; i<numIn; i+=1 {
		inType := m.Type().In(i)
		if inType==reflect.TypeOf([]string{}) {
			// command line parameters
			params[i] = reflect.ValueOf(os.Args[1:])
		} else if (inType.Kind()==reflect.Struct) || (inType.Kind()==reflect.Ptr && inType.Elem().Kind()==reflect.Struct)  {
			// command local configurations
			var conf reflect.Value
			if inType.Kind()==reflect.Ptr {
				conf = reflect.New(inType.Elem())
			} else {
				conf = reflect.New(inType)
			}
			file := AppConfDir() + "/" + cmd + ".toml"
			_,err := toml.DecodeFile(file, conf.Interface())
			if err != nil {
				return fmt.Errorf("Parse config [%s] failed: %s", file, err.Error())
			}
			if inType.Kind()==reflect.Ptr {
				params[i] = conf
			} else {
				params[i] = conf.Elem()
			}
		} else {
			return fmt.Errorf("No command [%s] because [cannot recognize func %s parameter %d]", cmd, mn, i)
		}
	}

	// global configurations
	if e:=parseConf(v); e!=nil {
		return e;
	}

	m.Call(params)

	return nil
}

func parseConf(v reflect.Value) error {

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return fmt.Errorf("AppStart parameter should be pointer to struct, %s given", v.Type().String())
	}

	// parse configuration file first
	for i:=0; i< e.NumField(); i+=1 {

		f :=  e.Field(i)
		if !f.IsValid() || !f.CanSet() || f.Kind()!=reflect.Struct {
			continue
		}

		name := e.Type().Field(i).Name
		if strings.HasPrefix(name,"Conf") {
			conf := utils.LcFirst(strings.TrimPrefix(name,"Conf"))
			conf = strings.TrimPrefix(conf,"_")
			file := AppConfDir() + "/" + conf + ".toml"
			_,err := toml.DecodeFile(file,f.Addr().Interface())
			if err!=nil {
				return fmt.Errorf("field [%s] parse configuration failed: %s", name, err.Error())
			}
		}

		// tag := e.Type().Field(i).Tag
		// fmt.Printf("field [%s] tag[%s]", e.Type().Field(i).Name,tag)
	}

	return nil
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
