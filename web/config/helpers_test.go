package webconfig

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const (
	RsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQBlDWBWg1VfPg2yhK9b25mgH+IFdi0p3NB7aV8ie3HXC7sHahor
Lly4sbU8tkw4Khz/QVkgoEXFhamBdq8ftlGox6/1XKQ4jIMUnsmA5+sr/EdWCWht
aLhjvwCP0+VyZRK9cyYq97HYfjn/AOWZodNkTCWP8QLZPMst1Lqah9kFQ6dXcEc2
7pG7+dWDMYB3Ssj9KGKJADnaNbDSvG6wzxcR7pmV1F3V32bnY0lwkP56dQDRImQf
lLkfM5WtziO+HDpf/lZSIP+PojIzVtgWHunvRlFWq45s3pVCr2k8gql9e91KymUg
twctCjIORLG6tZlJWheZATTOo/RjYnuSt3GBAgMBAAECggEAY4nAipdFgdKTFXSD
c10lv0eIkM6ugBzlM5tAvauRID2K1ivaaQx7AubxWkNLYsuA0qJhE7Ur5PvbUw5A
RGGHoFRKa8r8bENkXuxmGlbYHPScpBASUCsl7gLXn9UenwcWZYzGXO6tvU3g/OG6
SBHSM2qaWWYS4z9x98win1MPH5IeL4qG4aP8p9UUeszlZ3jRCgbEE5g/s4ZlOpk6
mePPbPAbNVSHVHWm8vpjaj5PyyJBlo79KifOa6UoC35WpjY2N+XKRskc5J5CYG3W
8s+eI83sRCRfaLpW3oMSWf5n7Ck6pR1S0z9CKIE4icu8/mEZK5sMDDreVsetj0Gj
8LEsbQKBgQCp15rfZlzdEOpqeC/9AxUJUJ/dxCAFa4qNpESfLli//iV1Z7AeS2xA
IdDs5D2wUMMGcrhfaZo/+SQJGJM9WZW7XGrsuISy+H3T48MGtYgM/SXYzIZkT6yI
pbKk/21+eH7ScxiPNklc6pOCg0uttXN+cURK7CwdLTtid4f0zIDA1wKBgQCYUG4K
SwYWWY9fDSxxwwugwTTvXwSOb8ta4u/qjv71SUytFE6C1VQbZl0vYZznJ4SVstPq
y4YQfY5Bb51fmnVOt2Nzo4DRmyGAVD9/46yeXOKCBm3I2FuC5mozvEbSYL6AYMAR
hEAH+9q6318opX4I8s34xoy5wAHOyy8izcqdZwKBgCF28vQLpkaEm7uovMPchJuW
QDmg+DI5M06fTEoVBrWlrNnKOWBuugCU6SaT/jesYhD2L6/VDeiG/XuxVOmkqiUF
m6fS34qn/ZrvaLD8ubbddkSWX/CJ5lPyWhW6omdl9I8g0Hj/WEsviSsRFas/cmc6
iBs/F54UGHxZUveAB+p3AoGBAJYbzhb+0UT9hrqadQ64tqD7SHwQHKFzs+oidWeo
/cQnNRXuSWHPdFB8laG4hbWZ0KTn1Mo1Qi7GTohWw7YQt0OYoL66iJ4KUP30CZC0
hM6do+2ZAGFMOQPMqSXLi9CR360NrLyWpbQ4ldxNf67DSm7vO+bNZqs5e46d2Bpo
ziJJAoGAev0+w2KgGKLRCOCeOUI8gQN3Lrszfh1P+0aPkYdawDhz2E2wcpoxasTz
EPyrPSXPxV5CVt2oIgVSu/x0bBFZWP09zduxxn8j3kFhKYYulVUOWSruSS9x7GvR
62k2mGlqKYAbUfkIEHi2PdR67jX8GH2DpRDFH4jbczLYENWaMKw=
-----END RSA PRIVATE KEY-----`

	RsaPubKey = `-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBlDWBWg1VfPg2yhK9b25mg
H+IFdi0p3NB7aV8ie3HXC7sHahorLly4sbU8tkw4Khz/QVkgoEXFhamBdq8ftlGo
x6/1XKQ4jIMUnsmA5+sr/EdWCWhtaLhjvwCP0+VyZRK9cyYq97HYfjn/AOWZodNk
TCWP8QLZPMst1Lqah9kFQ6dXcEc27pG7+dWDMYB3Ssj9KGKJADnaNbDSvG6wzxcR
7pmV1F3V32bnY0lwkP56dQDRImQflLkfM5WtziO+HDpf/lZSIP+PojIzVtgWHunv
RlFWq45s3pVCr2k8gql9e91KymUgtwctCjIORLG6tZlJWheZATTOo/RjYnuSt3GB
AgMBAAE=
-----END PUBLIC KEY-----`
)

func tmpFileSetup(t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "*GoTest.txt")

	if err != nil {
		t.Errorf("unable to create the tmp file [%s]", err.Error())
		t.SkipNow()
	}
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	return tmpFile.Name()
}

func tmpDirSetup(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "*GoTest")

	if err != nil {
		t.Errorf("unable to create the tmp file [%s]", err.Error())
		t.SkipNow()
	}
	t.Cleanup(func() {
		os.Remove(tmpDir)
	})
	return tmpDir
}

func TestIpValid(t *testing.T) {
	ips := [...]string{"127.0.0.1", "::1", "0.0.0.0"}
	for _, ip := range ips {
		got := CheckIfIpValid(ip)
		if got != nil {
			t.Errorf("ip validation failed for [%s]", ip)
		}
	}
}

func TestIpInvalid(t *testing.T) {
	ips := [...]string{"1.2.3", "333", "loclhost"}

	for _, ip := range ips {
		err := CheckIfIpValid(ip)
		if err == nil {
			t.Errorf("ip invalidation failed for [%s]", ip)
			return
		} else if !strings.Contains(err.Error(), "is not a valid IP") {
			t.Error("incorrect error message used")
			return
		}
	}

	// use speciall case to the localhost address
	err := CheckIfIpValid("localhost")
	if err == nil {
		t.Error("this is not a valid ip")
	} else if !errors.Is(err, ErrLocalhostNotValidIp) {
		t.Error("wrong error message used here")
	}
}

func TestPortValid(t *testing.T) {
	allowed := [...]int{-1}
	ports := [...]string{"111", "-1", "666", "66"}
	for _, port := range ports {
		err := CheckIfPortValid(port, allowed[:])
		if err != nil {
			t.Errorf("port validation failed for [%s]", port)
		}
	}
}

func TestPortInvalid(t *testing.T) {
	allowed := [...]int{-1}
	check := func(subStr string, ports []string) bool {
		for _, port := range ports {
			err := CheckIfPortValid(port, allowed[:])
			if err == nil {
				t.Errorf("port invalidation failed for [%s]", port)
				return false
			} else if !strings.Contains(err.Error(), subStr) {
				t.Error("wrong error message used")
				return false
			}
		}
		return true
	}
	port := []string{"-5", "-666"}
	if !check("Invalid value", port) {
		return
	}

	port = []string{"aa"}
	if !check("bad port value", port) {
		return
	}
}

func TestRSAValidatorValid(t *testing.T) {
	priKeyFile := tmpFileSetup(t)
	err := ioutil.WriteFile(priKeyFile, []byte(RsaPrivateKey), 0666)
	if err != nil {
		t.Error("unable to write to the tmp file")
		return
	}

	err = CheckIfRSAValid(priKeyFile)
	if err != nil {
		t.Error("there should not be an error here")
	}
}

func TestRSAValidatorEmptryFile(t *testing.T) {
	// check empty, but valid file
	priKeyFile := tmpFileSetup(t)
	err := CheckIfRSAValid(priKeyFile)

	if err == nil {
		t.Error("there should not be an error here")
	} else if !errors.Is(err, InvalidRSAKeyPath{}) {
		t.Error("incorrect rsa key")
	} else if !strings.Contains(err.Error(), "is not a valid RSA private key") {
		t.Error("incorrect rsa error message")
	}
}

func TestRSAValidatorNotAPath(t *testing.T) {
	// Attention this test does not test anything
	// of value, it only is for coverage
	orgAbs := getAbsPath
	getAbsPath = func(path string) (string, error) {
		return "", errors.New("some weird issue")
	}
	t.Cleanup(func() {
		getAbsPath = orgAbs
	})

	err := CheckIfRSAValid(".")

	if err == nil {
		t.Error("there should not be an error here")
	}
}

func TestRSAValidatorNoFile(t *testing.T) {
	priKeyDir := tmpDirSetup(t)
	// make sure it does not exist
	priKeyFile := priKeyDir + "/SomeNotExistingFile.txt"
	err := CheckIfRSAValid(priKeyFile)

	if err == nil {
		t.Error("there should not be an error here")
	} else if !errors.Is(err, InvalidRSAKeyPath{}) {
		t.Error("incorrect rsa key")
	} else if !strings.Contains(err.Error(), "is not a valid RSA private key") {
		t.Error("incorrect rsa error message")
	}
}

func TestRSAValidatorIsDirectory(t *testing.T) {
	priKeyDir := tmpDirSetup(t)
	err := CheckIfRSAValid(priKeyDir)

	if err == nil {
		t.Error("there should not be an error here")
	} else if !errors.Is(err, InvalidRSAKeyPath{}) {
		t.Error("incorrect rsa key")
	} else if !strings.Contains(err.Error(), "is not a valid RSA private key") {
		t.Error("incorrect rsa error message")
	} else if !errors.Is(errors.Unwrap(err), ErrIsDirectory) {
		t.Error("it is a folder and the error should say that")
	}
}
