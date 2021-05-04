package setup

import (
	"errors"
	"net"
	"testing"
)

type ConfigHelperOuterValid struct {
	Ip    string
	Inner ConfigHelperInnerValid
}

func (c *ConfigHelperOuterValid) CheckConfiguration() error {
	if net.ParseIP(c.Ip) == nil {
		return errors.New("some")
	}
	return nil
}

type ConfigHelperInnerValid struct {
	Num        int
	val        *int
	AnOtherVal *int
}

func (c ConfigHelperInnerValid) CheckConfiguration() error {
	if c.Num > 0 {
		return nil
	}

	return errors.New("incorrect value for num")
}

type ConfigHelperInnerInvalid struct {
	Num int
}

func TestCheckIfConfigValidStruct(t *testing.T) {
	type ConfigHelperOuterValid struct {
		Web ConfigHelperInnerValid
	}
	val := 42
	cho := ConfigHelperOuterValid{
		Web: ConfigHelperInnerValid{
			42,
			&val,
			&val,
		},
	}

	err := checkIfConfigRecurse(cho)

	if err != nil {
		t.Errorf("There should not be an error here [%s]", err.Error())
	}
}

func TestCheckIfConfigInvalidStruct(t *testing.T) {
	type ConfigHelperOuterValid struct {
		Web ConfigHelperInnerValid
	}
	val := 42
	cho := ConfigHelperOuterValid{
		Web: ConfigHelperInnerValid{
			0,
			&val,
			&val,
		},
	}

	err := checkIfConfigRecurse(cho)

	if err == nil {
		t.Errorf("There should be an error here")
	}
}

func TestCheckIfConfigValidPtr(t *testing.T) {
	type Config struct {
		Web *ConfigHelperInnerValid
	}
	val := 33
	cho := Config{
		Web: &ConfigHelperInnerValid{
			42,
			&val,
			&val,
		},
	}

	err := checkIfConfigRecurse(cho)

	if err != nil {
		t.Errorf("There should not be an error here [%s]", err.Error())
	}
}

func TestCheckIfConfigNilStruct(t *testing.T) {
	err := checkIfConfigRecurse(nil)
	if err != nil {
		t.Error("there should not be an error for this case")
	}
}

func TestCheckIfConfigNilPtr(t *testing.T) {
	var conf *ConfigHelperInnerValid
	err := checkIfConfigRecurse(conf)
	if err != nil {
		t.Error("there should not be an error for this case")
	}
}

func TestCheckIfConfigPanic(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Error("there should be a panic here")
		}
	}()

	type Conf struct {
		Web *ConfigHelperInnerInvalid
	}

	cho := Conf{
		Web: &ConfigHelperInnerInvalid{
			42,
		},
	}

	_ = checkIfConfigRecurse(cho)
}

func TestCheckIfConfigRecurseValid(t *testing.T) {
	type Config struct {
		Outer ConfigHelperOuterValid
	}
	val := 42
	cho := Config{
		Outer: ConfigHelperOuterValid{
			Ip: "127.0.0.1",
			Inner: ConfigHelperInnerValid{
				42,
				&val,
				&val,
			},
		},
	}

	err := checkIfConfigRecurse(cho)

	if err != nil {
		t.Errorf("There should not be an error here [%s]", err.Error())
	}
}

func TestCheckIfConfigRecurseInValuid(t *testing.T) {
	type Config struct {
		Outer ConfigHelperOuterValid
	}
	val := 42
	cho := Config{
		Outer: ConfigHelperOuterValid{
			Ip: "127.0.0.1",
			Inner: ConfigHelperInnerValid{
				0,
				&val,
				&val,
			},
		},
	}

	err := checkIfConfigRecurse(cho)

	if err == nil {
		t.Errorf("There should be an error here")
	}
}
