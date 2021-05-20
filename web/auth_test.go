package web

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

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

func TestAuthTokenIssueing(t *testing.T) {

	// generate key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate private key: %s\n", err)
		return
	}

	var payload []byte
	// Create signed payload
	token := jwt.New()
	token.Set(`foo`, `bar`)
	payload, err = jwt.Sign(token, jwa.RS256, privKey)
	if err != nil {
		t.Errorf("failed to generate signed payload: %s\n", err)
		return
	}

	{ // Parse signed payload, and perform (1) verification of the signature
		// and (2) validation of the JWT token
		// Validation can be performed in a separate step using `jwt.Validate`
		token, err := jwt.Parse(
			payload,
			jwt.WithValidate(true),
			jwt.WithVerify(jwa.RS256, &privKey.PublicKey),
		)
		if err != nil {
			t.Errorf("failed to parse JWT token: %s\n", err)
			return
		}
		buf, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			t.Errorf("failed to generate JSON: %s\n", err)
			return
		}
		t.Logf("%s", buf)
	}
}

func TestAuthRequestWithValidJson(t *testing.T) {
	check := func(auth authenticationManager) {
		postBody, _ := json.Marshal(map[string]string{
			"username": "Phil",
		})
		req := httptest.NewRequest("POST", "/auth", bytes.NewBuffer(postBody))
		rr := httptest.NewRecorder()

		router := mux.NewRouter()

		router.HandleFunc("/auth", auth.AuthConnection())
		router.ServeHTTP(rr, req)
		res := rr.Result()
		if res.StatusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				res.StatusCode, http.StatusOK)
			t.SkipNow()
		}
	}

	auth, _ := newAuthManager("")
	check(auth)

	// write to tmp file
	tmpFile, err := ioutil.TempFile("", "*rsaKey")
	if err != nil {
		t.Errorf("unexpected error happend [%s]", err.Error())
		t.SkipNow()
	}

	err = ioutil.WriteFile(tmpFile.Name(), []byte(RsaPrivateKey), 0666)
	if err != nil {
		t.Errorf("unexpected error happend [%s]", err.Error())
		t.SkipNow()
	}

	auth, err = newAuthManager(tmpFile.Name())
	if err != nil {
		t.Errorf("unexpected error happend [%s]", err.Error())
		t.SkipNow()
	}
	check(auth)
}

func TestAuthTokenResponseVaild(t *testing.T) {
	postBody, _ := json.Marshal(map[string]string{
		"username": "Phil",
	})
	req := httptest.NewRequest("POST", "/auth", bytes.NewBuffer(postBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	auth, _ := newAuthManager("")

	router.HandleFunc("/auth", auth.AuthConnection())
	router.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			res.StatusCode, http.StatusOK)
	}

	if res.Header.Get("Set-Cookie") == "" {
		t.Errorf("did not receive JWT via Set-Cookie")
	}
}

func TestValidUsernameHtmlEscape(t *testing.T) {
	username := "<SCRIPT SRC=http://xss.rocks/xss.js></SCRIPT>"
	escaped, _ := validUsername(username)
	expected := "&lt;SCRIPT SRC=http://xss.rocks/xss.js&gt;&lt;/SCRIPT&gt;"
	if strings.Compare(escaped, expected) != 0 {
		t.Errorf("expected %s, got %s", expected, escaped)
	}
}
