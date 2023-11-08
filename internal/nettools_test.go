package internal

import "testing"

func TestGenerateIP(t *testing.T) {
	ip, err := GenerateIP("10.0.0.1/16", []string{"10.0.0.1"})
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
	ip, err = GenerateIP("fd42:5e5e:b701:bae5::1/64", []string{"fd42:5e5e:b701:bae5::1"})
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
	useIp := []string{}
	for {
		ip, err = GenerateIP("fd42:5e5e:b701:bae5::1/112", useIp)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(ip)
		useIp = append(useIp, ip)
	}
}
