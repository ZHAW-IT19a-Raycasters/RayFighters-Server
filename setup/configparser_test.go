package setup

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.zhaw.ch/raycasters/server/room"

	"github.com/pelletier/go-toml"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const fileName = "conf.toml"

const RsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
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

func compareConfigs(exp, got Configuration, t *testing.T) bool {
	if got.Web.Ip != exp.Web.Ip {
		t.Error("The parsed ip is incorrect")
		return false
	}

	if got.Web.Port != exp.Web.Port {
		t.Error("The parsed port is incorrect")
		return false
	}

	return true
}

func tmpFileSetup(t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "*")

	if err != nil {
		t.Errorf("unable to create the tmp file [%s]", err.Error())
		t.SkipNow()
	}
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	return tmpFile.Name()
}

func writeRSAFile(t *testing.T) string {
	tmpFile := tmpFileSetup(t)
	err := os.WriteFile(tmpFile, []byte(RsaPrivateKey), 0666)
	if err != nil {
		t.Error("unable to write tmp file")
		t.SkipNow()
	}
	return tmpFile
}

func newConfigurationLegal(t *testing.T) Configuration {
	rsaFile := writeRSAFile(t)
	logFile := tmpFileSetup(t)
	mapFile := tmpFileSetup(t)
	os.WriteFile(mapFile, []byte("[]"), 0666)
	return Configuration{
		MapConfig: room.MapConfig{
			JsonPath: mapFile,
		},
		Web: webconfig.WebConfig{
			Ip:   "127.0.0.1",
			Port: 44,
			JWTConfig: webconfig.JwtConfig{
				RsaPrivateKey: rsaFile,
			},
			CORS: webconfig.CORSConfig{
				Domains: []string{
					// this system
					"http://127.0.0.1:8080",
					"http://localhost:8080",
					// frontend
					"http://127.0.0.1:9000",
					"http://localhost:9000",
				},
			},
		},
		Logger: LoggerConfig{
			Console: ConsoleConfig{
				Level: "debug",
			},
			File: FileConfig{
				Level:    "debug",
				FilePath: logFile,
			},
		},
	}
}

func marshalConfiguration(conf Configuration) ([]byte, error) {
	return toml.Marshal(conf)
}

func TestParseConfigLegal(t *testing.T) {
	rawConf := newConfigurationLegal(t)

	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling on control failed\n%s", err.Error())
		return
	}

	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}

	conf, err := confParser.Parse()

	if err != nil {
		t.Errorf("There should not be an error on this configuration file.\n%s", err.Error())
		return
	} else if conf == nil {
		t.Error("There should be a valid configuration file")
		return
	}

	compareConfigs(rawConf, *conf, t)
}

func TestParseConfigOsErr(t *testing.T) {
	// Any of these errors are unlikely as per definition
	// they only serve to get a higher coverage
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return nil, os.ErrNotExist
	}

	conf, err := confParser.Parse()

	if err == nil {
		t.Error("There should be an OS Error here")
		return
	}

	// as configurationError is a error wrapper, this is important
	if !errors.Is(err, os.ErrNotExist) {
		// this is an impossible case
		t.Error("Wrong error given")
		return
	}

	if !strings.HasPrefix(err.Error(), "during parsing of the configuration file happend") {
		t.Error("incorrect prefix used for error")
		return
	}

	if conf != nil {
		// this is an impossible case
		t.Error("a configuration was returned")
		return
	}
}

func TestParserIncorrectConfig(t *testing.T) {
	rawConf := newConfigurationLegal(t)
	// using an illigal ip
	rawConf.Web.Ip = "localho"

	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling failed\n%s", err.Error())
		return
	}

	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}
	conf, err := confParser.Parse()

	if conf != nil {
		t.Error("there should not be a configuration returned here.")
		return
	}

	if err == nil {
		t.Error("an error should have been used here")
		return
	}

	if !errors.Is(err, ErrConfiguration{}) {
		t.Error("incorrect error has been used")
		return
	}
}

func TestParserIlligalConfig(t *testing.T) {

	confData := []byte("a random string")
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}

	conf, err := confParser.Parse()

	if conf != nil {
		t.Error("there should not be a configuration returned here.")
		return
	}

	if err == nil {
		t.Error("an error should have been used here")
		return
	}

	if !errors.Is(err, ErrConfiguration{}) {
		t.Error("incorrect error has been used")
		return
	}
}
