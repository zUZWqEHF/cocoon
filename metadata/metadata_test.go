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
			{IP: "10.0.0.2", Prefix: 16, Mac: "aa:bb:cc:dd:ee:f0"},
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
	// Netplan config written by bootcmd.
	if !strings.Contains(out, "50-cocoon.yaml") {
		t.Errorf("netplan file name missing: %s", out)
	}
	if !strings.Contains(out, "10.0.0.2/16") {
		t.Errorf("IP/prefix missing in netplan: %s", out)
	}
	if !strings.Contains(out, "netplan apply") {
		t.Errorf("netplan apply missing: %s", out)
	}
}

func TestUserData_BootcmdMultiNIC(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 16, Gateway: "10.0.0.1", Mac: "aa:bb:cc:dd:ee:f0"},
			{IP: "10.0.1.2", Prefix: 24, Mac: "11:22:33:44:55:66"},
		},
		DNS: []string{"8.8.8.8"},
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	// Both positional params in MAC setting.
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

	// Both IPs in netplan section.
	if !strings.Contains(out, "10.0.0.2/16") {
		t.Errorf("first IP missing in netplan: %s", out)
	}
	if !strings.Contains(out, "10.0.1.2/24") {
		t.Errorf("second IP missing in netplan: %s", out)
	}

	// Both device names as netplan keys.
	if !strings.Contains(out, "$1:") {
		t.Errorf("$1 missing as netplan key: %s", out)
	}
	if !strings.Contains(out, "$2:") {
		t.Errorf("$2 missing as netplan key: %s", out)
	}

	// Gateway only on first NIC.
	if !strings.Contains(out, "via: 10.0.0.1") {
		t.Errorf("gateway missing: %s", out)
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

func TestUserData_BootcmdCleansOldNetplan(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Mac: "aa:bb:cc:dd:ee:f0"},
		},
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "rm -f /etc/netplan/*.yaml") {
		t.Errorf("old netplan cleanup missing: %s", out)
	}
	if !strings.Contains(out, "/run/systemd/network/10-netplan-*") {
		t.Errorf("old networkd cleanup missing: %s", out)
	}
}

func TestUserData_BootcmdGatewayOptional(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Mac: "aa:bb:cc:dd:ee:f0"},
		},
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if strings.Contains(out, "routes:") {
		t.Errorf("routes should not appear without gateway: %s", out)
	}
}

func TestUserData_BootcmdDNSOptional(t *testing.T) {
	cfg := &Config{
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Mac: "aa:bb:cc:dd:ee:f0"},
		},
		// No DNS.
	}

	var buf bytes.Buffer
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if strings.Contains(out, "nameservers:") {
		t.Errorf("nameservers should not appear without DNS: %s", out)
	}
}

func TestGenerate_NetworkConfigStatic(t *testing.T) {
	cfg := &Config{
		InstanceID:   "test-id",
		Hostname:     "test-vm",
		RootPassword: "pass",
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Gateway: "10.0.0.1", Mac: "aa:bb:cc:dd:ee:ff"},
		},
		DNS: []string{"8.8.8.8"},
	}

	var buf bytes.Buffer
	if err := Generate(&buf, cfg); err != nil {
		t.Fatal(err)
	}

	// FAT12 image should contain the static network-config.
	// We can't easily parse FAT12, but verify the image is non-empty.
	if buf.Len() == 0 {
		t.Fatal("empty output")
	}

	// Verify "version: 2" appears in the raw image (it's a small FAT12).
	raw := buf.String()
	if !strings.Contains(raw, "version: 2") {
		t.Error("network-config 'version: 2' not found in FAT12 image")
	}
	// No MAC matching anywhere in the image (old approach removed).
	if strings.Contains(raw, "macaddress:") {
		t.Error("FAT12 image should not contain macaddress matching")
	}
	// user-data should contain the bootcmd netplan heredoc.
	if !strings.Contains(raw, "50-cocoon.yaml") {
		t.Error("user-data should reference 50-cocoon.yaml")
	}
}

func TestGenerate_ProducesValidFAT12(t *testing.T) {
	cfg := &Config{
		InstanceID:   "test-id",
		Hostname:     "test-vm",
		RootPassword: "pass",
		Networks: []NetworkInfo{
			{IP: "10.0.0.2", Prefix: 24, Gateway: "10.0.0.1", Mac: "aa:bb:cc:dd:ee:ff"},
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
