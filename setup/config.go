package setup

import (
	"fmt"
	"reflect"

	"github.zhaw.ch/raycasters/server/room"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const (
	// Is the path to the default configuration file.
	Config = "config.toml"
)

// Represents a collection of configuration, that is read out of the
// configuration file.
// As per definition if there is a config file in here
// It must implement the CheckConfig interface.
type Configuration struct {
	MapConfig room.MapConfig      `toml:"Map"`
	Web       webconfig.WebConfig `toml:"Webconfig"`
	Logger    LoggerConfig        `toml:"Logging"`
}

// CheckIfConfigIsLegal will check that the configurated values are legal
// Attention this function uses reflection in a recursive loop and will
// panic if any struct inside of configuration do not implement
// the CheckConfig interface.
func CheckIfConfigIsLegal(conf *Configuration) error {
	// as config it self does not implement the check
	// ignore that part
	return checkIfConfigRecurse(conf)
}

// checkIfConfigRunCheck will do the actual check for the correctness state
// of the toCheck value.
func checkIfConfigRunCheck(toCheck CheckConfig, valRaw reflect.Value) error {
	// run the check and respond to it
	err := toCheck.CheckConfiguration()
	if err != nil {
		return &ErrConfiguration{
			section:     valRaw.Type().Name(),
			internalErr: err,
		}
	}
	return nil
}

// checkIfConfigCasePtr will check the correctness of a pointer type
func checkIfConfigCasePtr(valRaw reflect.Value, val interface{}) error {
	// pointer of an unwanted type
	if valRaw.Elem().Kind() != reflect.Struct {
		return nil
	}
	// check if the struct does implement
	// the given check function
	checkAble, ok := val.(CheckConfig)

	// case that maybe the struct itself
	// implements the CheckConfig
	// interface instead.
	if !ok {
		elem := valRaw.Elem()
		return checkIfConfigCaseStruct(elem, val)
	}

	return checkIfConfigRunCheck(checkAble, valRaw)
}

// checkIfConfigCaseStruct will check the correctness of a struct type
func checkIfConfigCaseStruct(valRaw reflect.Value, valInt interface{}) error {
	// check if the struct does implement
	// the given check function
	checkAble, ok := valInt.(CheckConfig)

	if ok {
		return checkIfConfigRunCheck(checkAble, valRaw)
	}

	// try it with the pointer version
	// do some black magic to get a pointer version

	// first make a copy of the object type (fake pointer)
	valRawPtr := reflect.New(reflect.TypeOf(valInt))
	// second add to the object the sub object
	// (makes a pointer out of the raw object type)
	valRawPtr.Elem().Set(reflect.ValueOf(valInt))

	valIntPtr := valRawPtr.Interface()
	checkAble, ok = valIntPtr.(CheckConfig)

	// neither the pointer nor the struct implement
	// the CheckConfig interface so this has to
	// crash.
	if !ok {
		str := fmt.Sprintf("implemented struct field [%s] does not implement `CheckConfig`", valRaw.Type().Name())
		panic(str)
	}

	return checkIfConfigRunCheck(checkAble, valRaw)
}

// checkIfConfigHandleElement will handle the actual reflected values
// so that they can be processed as required.
func checkIfConfigHandleElement(valRaw reflect.Value) error {

	// check if exported else => move on
	if !valRaw.CanInterface() {
		return nil
	}
	// get raw interface to the value
	valInt := valRaw.Interface()

	// check if value is a struct or not
	kind := valRaw.Kind()

	var err error
	switch kind {
	case reflect.Struct:
		err = checkIfConfigCaseStruct(valRaw, valInt)
	case reflect.Ptr:
		err = checkIfConfigCasePtr(valRaw, valInt)
	default:
		// the value is some other uninteresting type
		return nil
	}

	if err != nil {
		return err
	}

	// recurs over the checked value
	return checkIfConfigRecurse(valInt)
}

// checkIfConfigRecurse will check if the values
func checkIfConfigRecurse(val interface{}) error {

	// prepare the system
	raw := reflect.ValueOf(val)

	// do not work with pointers
	if raw.Kind() == reflect.Ptr {
		if raw.IsNil() {
			return nil
		}
		raw = raw.Elem()
	}

	// only deal with sturcts
	if raw.Kind() != reflect.Struct {
		return nil
	}

	// loop over the internals
	for i := 0; i < raw.NumField(); i += 1 {
		// try to understand the given value.
		valRaw := raw.Field(i)

		err := checkIfConfigHandleElement(valRaw)
		if err != nil {
			return err
		}

	}
	return nil
}
