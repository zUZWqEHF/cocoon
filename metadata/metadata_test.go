package metadata

import (
	"bytes"
	"strings"
	"testing"
)

func TestUserData_BootcmdSetsMAC(t *testing.T) {
	cfg := &Config{
		RootPassword: "test",
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 16, Device: "eth0", Mac: "aa:bb:cc:dd:ee:f0"},
		},
		DNS: []string{"8.8.8.8"},
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "bootcmd:") {
		t.Fatal("bootcmd section missing")
	}
	if !strings.Contains(out, `address 'aa:bb:cc:dd:ee:f0'`) {
		t.Errorf("MAC not in bootcmd: %s", out)
	}
	// $1 = first non-lo interface.
	if !strings.Contains(out, `"$1"`) {
		t.Errorf("positional param $1 missing: %s", out)
	}
}

func TestUserData_BootcmdMultiNIC(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 16, Device: "eth0", Mac: "aa:bb:cc:dd:ee:f0"},
			{IP: "10.0.1.2", Prefix: 24, Device: "eth1", Mac: "11:22:33:44:55:66"},
		},
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, `"$1"`) {
		t.Errorf("$1 missing: %s", out)
	}
	if !strings.Contains(out, `"$2"`) {
		t.Errorf("$2 missing: %s", out)
	}
	if !strings.Contains(out, `address 'aa:bb:cc:dd:ee:f0'`) {
		t.Errorf("first MAC missing: %s", out)
	}
	if !strings.Contains(out, `address '11:22:33:44:55:66'`) {
		t.Errorf("second MAC missing: %s", out)
	}
}

func TestUserData_NoNetworks(t *testing.T) {
	cfg := &Config{RootPassword: "test"}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if strings.Contains(out, "bootcmd:") {
		t.Errorf("bootcmd should not appear without networks: %s", out)
	}
}

func TestNetworkConfig_MACMatch(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.88.0.45", Prefix: 16, Gateway: "10.88.0.1", Device: "eth0", Mac: "fe:86:bf:1f:53:17"},
		},
		DNS: []string{"8.8.8.8", "1.1.1.1"},
	}

	var buf bytes.Buffer
	if err := networkConfigTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "macaddress: 'fe:86:bf:1f:53:17'") {
		t.Errorf("MAC match missing: %s", out)
	}
	if !strings.Contains(out, "10.88.0.45/16") {
		t.Errorf("IP missing: %s", out)
	}
}

func TestGenerate_ProducesValidFAT12(t *testing.T) {
	cfg := &Config{
		InstanceID:   "test-id",
		Hostname:     "test-vm",
		RootPassword: "pass",
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Gateway: "10.0.0.1", Device: "eth0", Mac: "aa:bb:cc:dd:ee:ff"},
		},
		DNS: []string{"8.8.8.8"},
	}

	var buf bytes.Buffer
	if err := Generate(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty output")
	}
}
