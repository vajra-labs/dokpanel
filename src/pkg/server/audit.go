package server

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"strings"

	"goploy/src/pkg/shellx"
)

type UfwAudit struct {
	Installed       bool   `json:"installed"`
	Active          bool   `json:"active"`
	DefaultIncoming string `json:"defaultIncoming"`
}

type SshAudit struct {
	Enabled         bool   `json:"enabled"`
	KeyAuth         bool   `json:"keyAuth"`
	PermitRootLogin string `json:"permitRootLogin"`
	PasswordAuth    string `json:"passwordAuth"`
	UsePam          string `json:"usePam"`
}

type Fail2banAudit struct {
	Installed  bool   `json:"installed"`
	Enabled    bool   `json:"enabled"`
	Active     bool   `json:"active"`
	SshEnabled string `json:"sshEnabled"`
	SshMode    string `json:"sshMode"`
}

type ServerAuditResult struct {
	Ufw      UfwAudit      `json:"ufw"`
	Ssh      SshAudit      `json:"ssh"`
	Fail2ban Fail2banAudit `json:"fail2ban"`
}

// Thanks for the idea to https://github.com/healthyhost/audit-vps-script/tree/main
const ufwCheck = `
  if command -v ufw >/dev/null 2>&1; then
    isInstalled=true
    isActive=$(sudo ufw status | grep -q "Status: active" && echo true || echo false)
    defaultIncoming=$(sudo ufw status verbose | grep "Default:" | grep "incoming" | awk '{print $2}')
    echo "{\"installed\": $isInstalled, \"active\": $isActive, \"defaultIncoming\": \"$defaultIncoming\"}"
  else
    echo "{\"installed\": false, \"active\": false, \"defaultIncoming\": \"unknown\"}"
  fi
`

const sshCheck = `
  if systemctl is-active --quiet sshd || systemctl is-active --quiet ssh; then
    isEnabled=true

    # Get the sshd config file path
    sshd_config=$(sudo sshd -T 2>/dev/null | grep -i "^configfile" | awk '{print $2}')
    
    # If we couldn't get the path, use the default
    if [ -z "$sshd_config" ]; then
      sshd_config="/etc/ssh/sshd_config"
    fi

    # Check for key authentication
    # SSH key auth is enabled by default unless explicitly disabled
    pubkey_line=$(sudo grep -i "^PubkeyAuthentication" "$sshd_config" 2>/dev/null | grep -v "#")
    if [ -z "$pubkey_line" ] || echo "$pubkey_line" | grep -q -i "yes"; then
      keyAuth=true
    else
      keyAuth=false
    fi

    # Get the exact PermitRootLogin value from config
    # This preserves values like "prohibit-password" without normalization
    permitRootLogin=$(sudo grep -i "^PermitRootLogin" "$sshd_config" 2>/dev/null | grep -v "#" | awk '{print $2}')
    if [ -z "$permitRootLogin" ]; then
      # Default is prohibit-password in newer versions
      permitRootLogin="prohibit-password"
    fi
    
    # Get the exact PasswordAuthentication value from config
    passwordAuth=$(sudo grep -i "^PasswordAuthentication" "$sshd_config" 2>/dev/null | grep -v "#" | awk '{print $2}')
    if [ -z "$passwordAuth" ]; then
      # Default is yes
      passwordAuth="yes"
    fi

    # Get the exact UsePAM value from config
    usePam=$(sudo grep -i "^UsePAM" "$sshd_config" 2>/dev/null | grep -v "#" | awk '{print $2}')
    if [ -z "$usePam" ]; then
      # Default is yes in most distros
      usePam="yes"
    fi

    # Return the results with exact values from config file
    echo "{\"enabled\": $isEnabled, \"keyAuth\": $keyAuth, \"permitRootLogin\": \"$permitRootLogin\", \"passwordAuth\": \"$passwordAuth\", \"usePam\": \"$usePam\"}"
  else
    echo "{\"enabled\": false, \"keyAuth\": false, \"permitRootLogin\": \"unknown\", \"passwordAuth\": \"unknown\", \"usePam\": \"unknown\"}"
  fi
`

const fail2banCheck = `
  if dpkg -l | grep -q "fail2ban"; then
    isInstalled=true
    isEnabled=$(systemctl is-enabled --quiet fail2ban.service && echo true || echo false)
    isActive=$(systemctl is-active --quiet fail2ban.service && echo true || echo false)

    if [ -f "/etc/fail2ban/jail.local" ]; then
      sshEnabled=$(grep -A10 "^\[sshd\]" /etc/fail2ban/jail.local | grep "enabled" | awk '{print $NF}' | tr -d '[:space:]')
      sshMode=$(grep -A10 "^\[sshd\]" /etc/fail2ban/jail.local | grep "^mode[[:space:]]*=[[:space:]]*aggressive" >/dev/null && echo "aggressive" || echo "normal")
      echo "{\"installed\": $isInstalled, \"enabled\": $isEnabled, \"active\": $isActive, \"sshEnabled\": \"$sshEnabled\", \"sshMode\": \"$sshMode\"}"
    else
      echo "{\"installed\": $isInstalled, \"enabled\": $isEnabled, \"active\": $isActive, \"sshEnabled\": \"false\", \"sshMode\": \"normal\"}"
    fi
  else
    echo "{\"installed\": false, \"enabled\": false, \"active\": false, \"sshEnabled\": \"false\", \"sshMode\": \"normal\"}"
  fi
`

// AuditServer executes a security audit on a remote server.
func AuditServer(
	ctx context.Context,
	pool *shellx.SSHPool,
	serverId int64,
) (*ServerAuditResult, error) {
	bashCommand := fmt.Sprintf(`
ufwStatus=$( %s )
sshStatus=$( %s )
fail2banStatus=$( %s )
echo "{\"ufw\": $ufwStatus, \"ssh\": $sshStatus, \"fail2ban\": $fail2banStatus}"
`, ufwCheck, sshCheck, fail2banCheck)
	// Execute via SSHPool
	resChan := pool.Exec(ctx, serverId, bashCommand, nil)
	res := <-resChan
	if res.Err != nil {
		return nil, res.Err
	}
	stdout := strings.TrimSpace(res.Stdout)
	var result ServerAuditResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, fmt.Errorf(
			"Failed to parse server audit output %q: %w",
			stdout,
			err,
		)
	}
	return &result, nil
}
