package sv

// 标准应用包的目录结构

import (
	"os"
	"fmt"
	"flag"
	"reflect"
	"strings"
	"strconv"
	"io/ioutil"
	"time"
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

func AppRun(app interface{}, args []string) error {

	v := reflect.ValueOf(app)
	if v.Kind()!=reflect.Ptr || v.IsNil() || v.Elem().Kind()!=reflect.Struct {
		return fmt.Errorf("AppRun parameter should be pointer to struct, %s given", v.Type().String())
	}

	// global flags
	if len(args)<1 {
		return fmt.Errorf("args should not be empty, expected format is 'cmd [-flag1] subcommand [-subflag1] ...'")
	}
	var subArgs []string
	if err:=AppParseFlags(args[0], app, args[1:], &subArgs); err!=nil {
		return fmt.Errorf("Parse flags failed\n%s",err.Error())
	}

	// global configurations
	if e:=AppParseConf(v.Interface()); e!=nil {
		return e;
	}

	// subcommand method
	if len(subArgs)<1 {
		return fmt.Errorf("You should specify a command to execute")
	}
	cmd := subArgs[0]
	mn := "Cmd"+utils.UcFirst(cmd)
	m := v.MethodByName(mn)
	if !m.IsValid() {
		return fmt.Errorf("No such command [%s]", cmd)
	}

	// subcommand method parameters
	numIn := m.Type().NumIn()
	params := make([]reflect.Value,numIn,numIn)
	if numIn>1 {
		return fmt.Errorf("No such command [%s] because [func %s declaration is invalid, too much parameters required]", cmd, mn)
	}
	for i:=0; i<numIn; i+=1 {
		// type check
		inType := m.Type().In(i)

		if inType == reflect.TypeOf(subArgs) {
			params[i] = reflect.ValueOf(subArgs)
			continue;
		}

		if (inType.Kind()!=reflect.Struct) && (inType.Kind()!=reflect.Ptr || inType.Elem().Kind()!=reflect.Struct)  {
			return fmt.Errorf("No command [%s] because [cannot recognize func %s parameter type]", cmd, mn)
		}

		// subcommand configurations and flags
		var options reflect.Value
		if inType.Kind()==reflect.Ptr {
			options = reflect.New(inType.Elem())
		} else {
			options = reflect.New(inType)
		}

		if err:=AppParseConf(options.Interface()); err!=nil {
			return fmt.Errorf("parse config failed! %s", err.Error())
		}

		if err:=AppParseFlags(cmd, options.Interface(),subArgs[1:],nil); err!=nil {
			return fmt.Errorf("parse flags failed! %s", err.Error())
		}

		if inType.Kind()==reflect.Ptr {
			params[i] = options
		} else {
			params[i] = options.Elem()
		}
	}

	m.Call(params)

	return nil
}

// @param i pointer to struct
func AppParseFlags(flagsetName string,i interface{}, flagArgs[]string, nonFlagArgs*[]string) error {

	flags := flag.NewFlagSet(flagsetName, flag.ContinueOnError)

	v := reflect.ValueOf(i)
	e := v.Elem()

	flagFieldCnt := 0
	// parse configuration file first
	for i:=0; i< e.NumField(); i+=1 {

		f :=  e.Field(i)
		if !f.IsValid() || !f.CanSet() {
			continue
		}

		fieldVal  := e.Field(i)
		fieldDesc := e.Type().Field(i)
		fieldName := fieldDesc.Name
		tag       := fieldDesc.Tag

		var flagName string     = ""
		var usage string        = ""
		var defaultStr string = ""

		if  tagType:=tag.Get("type"); (tagType!="flag") && !strings.HasPrefix(fieldName,"Flag") {
			continue
		}

		flagName = utils.LcFirst(strings.TrimPrefix(fieldName,"Flag"))
		flagName = strings.TrimPrefix(flagName,"_")
		if name:=tag.Get("name"); len(name)>0 {
			flagName = name
		}
		usage = tag.Get("usage")
		defaultStr = tag.Get("default")

		var err error = nil
		if fieldDesc.Type == reflect.TypeOf("") { // string
			flags.StringVar(fieldVal.Addr().Interface().(*string),flagName,defaultStr,usage)
		} else if fieldDesc.Type == reflect.TypeOf(true) { // bool
			defaultVal := false
			if len(defaultStr)>0 {
				defaultVal,err = strconv.ParseBool(defaultStr)
			}
			flags.BoolVar(fieldVal.Addr().Interface().(*bool),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(int(0)) { // int
			defaultVal := int(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := strconv.ParseInt(defaultStr,0,reflect.TypeOf(defaultVal).Bits())
				err = tmpErr
				defaultVal = int(tmp)
			}
			flags.IntVar(fieldVal.Addr().Interface().(*int),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(int64(0)) { // int64
			defaultVal := int64(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := strconv.ParseInt(defaultStr,0,reflect.TypeOf(defaultVal).Bits())
				err = tmpErr
				defaultVal = int64(tmp)
			}
			flags.Int64Var(fieldVal.Addr().Interface().(*int64),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(uint(0)) { // uint
			defaultVal := uint(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := strconv.ParseInt(defaultStr,0,reflect.TypeOf(defaultVal).Bits())
				err = tmpErr
				defaultVal = uint(tmp)
			}
			flags.UintVar(fieldVal.Addr().Interface().(*uint),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(uint64(0)) { // uint64
			defaultVal := uint64(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := strconv.ParseInt(defaultStr,0,reflect.TypeOf(defaultVal).Bits())
				err = tmpErr
				defaultVal = uint64(tmp)
			}
			flags.Uint64Var(fieldVal.Addr().Interface().(*uint64),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(float64(0)) { // float64
			defaultVal := float64(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := strconv.ParseFloat(defaultStr,reflect.TypeOf(defaultVal).Bits())
				err = tmpErr
				defaultVal = float64(tmp)
			}
			flags.Float64Var(fieldVal.Addr().Interface().(*float64),flagName,defaultVal,usage)
		} else if fieldDesc.Type == reflect.TypeOf(time.Duration(0)) { // duration
			defaultVal := time.Duration(0)
			if len(defaultStr)>0 {
				tmp,tmpErr := time.ParseDuration(defaultStr)
				err = tmpErr
				defaultVal = time.Duration(tmp)
			}
			flags.DurationVar(fieldVal.Addr().Interface().(*time.Duration),flagName,defaultVal,usage)
		} else {
			return fmt.Errorf("flag [%s] type[%s] is not supported yet",flagName,fieldDesc.Type.String())
		}
		if err!=nil {
			return fmt.Errorf("flag [%s] parse default value [%s] failed!",flagName, defaultStr)
		}

		flagFieldCnt += 1
	}

	argsField := e.FieldByName("Args")
	hasArgsField := argsField.IsValid() && !argsField.CanSet() && argsField.Type()==reflect.TypeOf([]string{})

	// no need to parse flags for this struct
	if flagFieldCnt==0 && nonFlagArgs==nil &&  !hasArgsField {
		return nil
	}

	err := flags.Parse(flagArgs)
	if err == nil {
		if nonFlagArgs != nil {
			*nonFlagArgs = flags.Args()
		}
		if hasArgsField {
			argsField.Set(reflect.ValueOf(flags.Args()))
		}
	}
	return err
}

// parse config file by reflection
func AppParseConf(i interface{}) error {

	v := reflect.ValueOf(i)
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return fmt.Errorf("Parameter should be pointer to struct, %s given", v.Type().String())
	}

	// parse configuration file first
	for i:=0; i< e.NumField(); i+=1 {

		f :=  e.Field(i)
		if !f.IsValid() || !f.CanSet() || f.Kind()!=reflect.Struct {
			continue
		}

		name := e.Type().Field(i).Name
		tag := e.Type().Field(i).Tag

		if src:=tag.Get("src"); len(src)!=0 {
			file := AppBaseDir() + "/" + src
			_,err := toml.DecodeFile(file,f.Addr().Interface())
			if err!=nil {
				return fmt.Errorf("field [%s] parse configuration failed: %s", name, err.Error())
			}
		} else if strings.HasPrefix(name,"Conf") {
			conf := strings.TrimPrefix(utils.LcFirst(strings.TrimPrefix(name,"Conf")),"_")
			file := AppConfDir() + "/" + conf + ".toml"
			_,err := toml.DecodeFile(file,f.Addr().Interface())
			if err!=nil {
				return fmt.Errorf("field [%s] parse configuration failed: %s", name, err.Error())
			}
		}

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
