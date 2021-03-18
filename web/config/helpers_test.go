package webconfig

import "testing"

func TestIpValid(t *testing.T) {
	ips := [...]string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
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
		got := CheckIfIpValid(ip)
		if got == nil {
			t.Errorf("ip invalidation failed for [%s]", ip)
		}
	}
}

func TestPortValid(t *testing.T) {
	allowed := [...]int{-1}
	ports := [...]string{"111", "-1", "666", "66"}
	for _, port := range ports {
		got := CheckIfPortValid(port, allowed[:])
		if got != nil {
			t.Errorf("port validation failed for [%s]", port)
		}
	}
}

func TestPortInvalid(t *testing.T) {
	allowed := [...]int{-1}
	ports := [...]string{"-5", "-666", "aa"}
	for _, port := range ports {
		got := CheckIfPortValid(port, allowed[:])
		if got == nil {
			t.Errorf("port invalidation failed for [%s]", port)
		}
	}
}
