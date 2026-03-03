package metadata

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

const cidataLabel = "CIDATA"

var (
	tmplFuncs = template.FuncMap{
		// yamlQuote escapes single quotes for YAML single-quoted strings.
		"yamlQuote": func(s string) string {
			return strings.ReplaceAll(s, "'", "''")
		},
		// add1 returns i+1, used to map 0-based index to shell positional params ($1, $2, …).
		"add1": func(i int) int {
			return i + 1
		},
	}

	metaDataTmpl = template.Must(template.New("meta-data").Parse(
		"instance-id: {{.InstanceID}}\nlocal-hostname: {{.Hostname}}\n"))

	// userDataTmpl renders cloud-config user-data.
	//
	// The bootcmd section does ALL networking in one shell script:
	//   1. Discover non-lo interfaces sorted alphabetically → $1, $2, …
	//   2. Set each NIC's MAC to the CNI-assigned address
	//   3. Clean stale netplan / systemd-networkd configs
	//   4. Write /etc/netplan/50-cocoon.yaml using $1/$2 device names (shell expansion)
	//   5. netplan apply
	//
	// This avoids netplan's match:macaddress → PermanentMACAddress pitfall entirely,
	// and works identically for create, clone-resume, and restart.
	userDataTmpl = template.Must(template.New("user-data").Funcs(tmplFuncs).Parse(`#cloud-config
{{- if .RootPassword}}
chpasswd:
  expire: false
  list:
    - 'root:{{yamlQuote .RootPassword}}'
ssh_pwauth: true
disable_root: false
{{- end}}
{{- if .Networks}}
bootcmd:
  - |
    set -- $(ls /sys/class/net/ | grep -v '^lo$' | sort){{range $i, $n := .Networks}}{{if $n.Mac}}
    [ -n "${{add1 $i}}" ] && ip link set dev "${{add1 $i}}" down && ip link set dev "${{add1 $i}}" address '{{$n.Mac}}' && ip link set dev "${{add1 $i}}" up || true{{end}}{{end}}
    rm -f /etc/netplan/*.yaml /run/systemd/network/10-netplan-*
    cat > /etc/netplan/50-cocoon.yaml << EONETPLAN
    network:
      version: 2
      ethernets:
{{- range $i, $n := .Networks}}
        ${{add1 $i}}:
          addresses:
            - {{$n.IP}}/{{$n.Prefix}}
{{- if $n.Gateway}}
          routes:
            - to: default
              via: {{$n.Gateway}}
{{- end}}
{{- if $.DNS}}
          nameservers:
            addresses:
{{- range $.DNS}}
              - {{.}}
{{- end}}
{{- end}}
{{- end}}
    EONETPLAN
    netplan apply
{{- end}}
`))
)

// Config holds the inputs for generating cloud-init NoCloud metadata.
type Config struct {
	InstanceID   string
	Hostname     string
	RootPassword string
	Networks     []NetworkInfo
	DNS          []string // e.g. ["8.8.8.8", "8.8.4.4"]
}

// NetworkInfo describes a single guest network interface for cloud-init.
type NetworkInfo struct {
	IP      string // e.g. "10.0.0.2"
	Prefix  int    // CIDR prefix length, e.g. 24
	Gateway string // e.g. "10.0.0.1"
	Mac     string // CNI-assigned MAC — used in bootcmd to set current MAC
}

// Generate streams a cloud-init NoCloud cidata disk image (FAT12) to w.
func Generate(w io.Writer, cfg *Config) error {
	files := make(map[string][]byte, 2) //nolint:mnd

	var buf bytes.Buffer
	if err := metaDataTmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("render meta-data: %w", err)
	}
	files["meta-data"] = bytes.Clone(buf.Bytes())

	buf.Reset()
	if err := userDataTmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("render user-data: %w", err)
	}
	files["user-data"] = bytes.Clone(buf.Bytes())

	// Static no-op network-config prevents cloud-init DHCP fallback.
	// Actual networking is configured by bootcmd (writes netplan + applies).
	if len(cfg.Networks) > 0 {
		files["network-config"] = []byte("version: 2\n")
	}

	return CreateFAT12(w, cidataLabel, files)
}
