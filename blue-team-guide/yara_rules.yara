/*
    Red Team C2 Framework - YARA Detection Rules

    PURPOSE: Detect C2 implants, teamservers, and related infrastructure
    USAGE: yara -r yara_rules.yara /path/to/scan

    AUTHOR: Red Team C2 Project
    DATE: 2024
    LICENSE: Educational Use Only
*/

// =====================================================================
// Rule 1: Detect Go-based HTTPS Beacon Implant
// =====================================================================

rule C2_HTTPS_Beacon_Implant
{
    meta:
        description = "Detects Go-based C2 implant with HTTPS beaconing"
        author = "Red Team C2 Project"
        date = "2024-11-16"
        severity = "critical"
        reference = "https://github.com/Raoof128/red-team-c2"

    strings:
        // Go runtime strings
        $go_runtime1 = "go.buildid" ascii
        $go_runtime2 = "runtime.main" ascii

        // Crypto library imports
        $crypto1 = "crypto/tls" ascii
        $crypto2 = "crypto/aes" ascii
        $crypto3 = "golang.org/x/crypto/chacha20poly1305" ascii

        // C2 specific strings
        $c2_beacon1 = "/api/beacon/" ascii
        $c2_beacon2 = "/api/commands/" ascii
        $c2_beacon3 = "beacon_id" ascii
        $c2_beacon4 = "BeaconConfig" ascii

        // HTTP client patterns
        $http1 = "User-Agent: Mozilla/" ascii
        $http2 = "Content-Type: application/json" ascii

        // Jitter/sleep patterns
        $jitter = "JitterPercent" ascii
        $interval = "BaseInterval" ascii

    condition:
        uint32(0) == 0x464c457f or  // ELF
        uint16(0) == 0x5a4d and     // PE
        (
            (2 of ($go_runtime*)) and
            (1 of ($crypto*)) and
            (2 of ($c2_beacon*))
        )
}

// =====================================================================
// Rule 2: Detect ChaCha20-Poly1305 Encryption Usage
// =====================================================================

rule C2_ChaCha20_Crypto
{
    meta:
        description = "Detects ChaCha20-Poly1305 AEAD cipher usage (common in C2)"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        // ChaCha20 constants
        $chacha_const1 = "expand 32-byte k" ascii
        $chacha_const2 = { 65 78 70 61 6E 64 20 33 32 2D 62 79 74 65 20 6B }

        // Poly1305 constants
        $poly1305 = "Poly1305" ascii nocase

        // Function names
        $func_encrypt = "Encrypt" ascii
        $func_decrypt = "Decrypt" ascii
        $func_aead = "AEAD" ascii

        // Library references
        $lib_chacha = "chacha20poly1305" ascii nocase

    condition:
        (1 of ($chacha_const*)) or
        (
            $lib_chacha and
            (1 of ($func_*))
        )
}

// =====================================================================
// Rule 3: Detect DNS Tunneling Implant
// =====================================================================

rule C2_DNS_Tunneling
{
    meta:
        description = "Detects DNS covert channel C2 implant"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        // DNS query patterns
        $dns1 = "LookupTXT" ascii
        $dns2 = "LookupHost" ascii
        $dns3 = "net.Resolver" ascii

        // Base32 encoding (common in DNS tunneling)
        $base32_1 = "base32.StdEncoding" ascii
        $base32_2 = "WithPadding" ascii

        // DNS-specific strings
        $dns_domain = "DNSDomain" ascii
        $dns_chunk = "chunkData" ascii
        $dns_beacon = "DNSBeacon" ascii

        // DNS resolvers
        $resolver1 = "8.8.8.8:53" ascii
        $resolver2 = "1.1.1.1:53" ascii

    condition:
        (2 of ($dns*)) and
        (1 of ($base32_*)) and
        (1 of ($resolver*))
}

// =====================================================================
// Rule 4: Detect Rust Teamserver
// =====================================================================

rule C2_Teamserver_Rust
{
    meta:
        description = "Detects Rust-based C2 teamserver"
        author = "Red Team C2 Project"
        severity = "critical"

    strings:
        // Rust runtime
        $rust1 = "rustc" ascii
        $rust2 = "cargo" ascii

        // Tokio async runtime
        $tokio1 = "tokio::runtime" ascii
        $tokio2 = "tokio::spawn" ascii

        // Axum web framework
        $axum1 = "axum::routing" ascii
        $axum2 = "axum::extract" ascii

        // PostgreSQL/SQLx
        $db1 = "sqlx::postgres" ascii
        $db2 = "PgPool" ascii

        // C2 endpoints
        $endpoint1 = "/api/beacon/" ascii
        $endpoint2 = "/api/operator/" ascii
        $endpoint3 = "/api/commands/" ascii

        // Database tables
        $table1 = "beacons" ascii
        $table2 = "commands" ascii
        $table3 = "outputs" ascii

    condition:
        (1 of ($rust*)) and
        (1 of ($tokio*) or 1 of ($axum*)) and
        (2 of ($endpoint*)) and
        (2 of ($table*))
}

// =====================================================================
// Rule 5: Detect Command Execution Module
// =====================================================================

rule C2_Command_Execution
{
    meta:
        description = "Detects C2 command execution capabilities"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        // Windows command execution
        $win_cmd1 = "cmd.exe" ascii wide
        $win_cmd2 = "powershell.exe" ascii wide
        $win_cmd3 = "ExecutionPolicy Bypass" ascii wide

        // Process creation
        $proc1 = "CreateProcess" ascii
        $proc2 = "exec.Command" ascii
        $proc3 = "os/exec" ascii

        // Output capture
        $output1 = "CombinedOutput" ascii
        $output2 = "StdoutPipe" ascii
        $output3 = "StderrPipe" ascii

        // Timeout handling
        $timeout = "context.WithTimeout" ascii

    condition:
        (2 of ($win_cmd*) or 2 of ($proc*)) and
        (1 of ($output*))
}

// =====================================================================
// Rule 6: Detect PowerShell Script Execution
// =====================================================================

rule C2_PowerShell_Execution
{
    meta:
        description = "Detects C2 PowerShell execution module"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        $ps1 = "powershell.exe -ExecutionPolicy Bypass" ascii wide
        $ps2 = "-NoProfile -NonInteractive" ascii wide
        $ps3 = "-Command" ascii wide
        $ps4 = "-EncodedCommand" ascii wide
        $ps5 = "-WindowStyle Hidden" ascii wide

        // AMSI bypass indicators
        $amsi1 = "AmsiScanBuffer" ascii wide
        $amsi2 = "AmsiInitFailed" ascii wide

        // ScriptBlock logging evasion
        $log_evasion = "ScriptBlock" ascii wide

    condition:
        (2 of ($ps*)) or
        (1 of ($amsi*))
}

// =====================================================================
// Rule 7: Detect In-Memory Execution
// =====================================================================

rule C2_InMemory_Execution
{
    meta:
        description = "Detects in-memory .NET assembly execution"
        author = "Red Team C2 Project"
        severity = "critical"

    strings:
        // .NET reflection
        $dotnet1 = "Assembly.Load" ascii wide
        $dotnet2 = "InvokeMember" ascii wide
        $dotnet3 = "GetMethod" ascii wide

        // Memory allocation
        $mem1 = "VirtualAlloc" ascii
        $mem2 = "VirtualProtect" ascii
        $mem3 = "WriteProcessMemory" ascii

        // Process injection
        $inject1 = "CreateRemoteThread" ascii
        $inject2 = "QueueUserAPC" ascii

    condition:
        (2 of ($dotnet*)) or
        (2 of ($mem*)) or
        (1 of ($inject*))
}

// =====================================================================
// Rule 8: Detect Persistence Mechanisms
// =====================================================================

rule C2_Persistence
{
    meta:
        description = "Detects C2 persistence mechanisms"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        // Registry Run keys
        $reg1 = "Software\\Microsoft\\Windows\\CurrentVersion\\Run" ascii wide
        $reg2 = "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run" ascii wide

        // Scheduled tasks
        $schtask1 = "schtasks /create" ascii wide
        $schtask2 = "Register-ScheduledTask" ascii wide

        // Services
        $service1 = "sc create" ascii wide
        $service2 = "New-Service" ascii wide

        // Startup folder
        $startup = "\\Start Menu\\Programs\\Startup" ascii wide

    condition:
        2 of them
}

// =====================================================================
// Rule 9: Detect Credential Harvesting
// =====================================================================

rule C2_Credential_Harvesting
{
    meta:
        description = "Detects credential dumping capabilities"
        author = "Red Team C2 Project"
        severity = "critical"

    strings:
        // LSASS memory dump
        $lsass1 = "lsass.exe" ascii wide nocase
        $lsass2 = "SeDebugPrivilege" ascii wide

        // Mimikatz strings
        $mimikatz1 = "sekurlsa::logonpasswords" ascii wide
        $mimikatz2 = "privilege::debug" ascii wide

        // SAM database
        $sam = "SAM\\Domains\\Account" ascii wide

        // Windows credential store
        $cred1 = "CredEnumerate" ascii
        $cred2 = "CredReadW" ascii

        // Browser credentials
        $browser1 = "Login Data" ascii
        $browser2 = "logins.json" ascii

    condition:
        2 of them
}

// =====================================================================
// Rule 10: Detect Lateral Movement
// =====================================================================

rule C2_Lateral_Movement
{
    meta:
        description = "Detects lateral movement capabilities"
        author = "Red Team C2 Project"
        severity = "high"

    strings:
        // Pass-the-hash
        $pth1 = "NTLMHash" ascii wide
        $pth2 = "SMBExec" ascii wide

        // WMI execution
        $wmi1 = "Win32_Process" ascii wide
        $wmi2 = "Create" ascii wide

        // PSExec-like
        $psexec1 = "\\\\%s\\ADMIN$" ascii wide
        $psexec2 = "IPC$" ascii wide

        // Kerberos
        $kerb1 = "Kerberoasting" ascii wide
        $kerb2 = "AS-REP" ascii wide

    condition:
        2 of them
}

// =====================================================================
// Rule 11: Detect Network Reconnaissance
// =====================================================================

rule C2_Network_Recon
{
    meta:
        description = "Detects network reconnaissance modules"
        author = "Red Team C2 Project"
        severity = "medium"

    strings:
        // Network scanning
        $scan1 = "net view" ascii wide
        $scan2 = "nltest /dclist" ascii wide
        $scan3 = "dsquery" ascii wide

        // Port scanning
        $port1 = "PortScan" ascii
        $port2 = "ConnectTimeout" ascii

        // AD enumeration
        $ad1 = "Get-ADUser" ascii wide
        $ad2 = "Get-ADGroup" ascii wide
        $ad3 = "Get-ADComputer" ascii wide

        // BloodHound
        $bloodhound = "SharpHound" ascii wide

    condition:
        2 of them
}

// =====================================================================
// Rule 12: Detect Python Operator CLI
// =====================================================================

rule C2_Operator_CLI
{
    meta:
        description = "Detects C2 operator Python CLI"
        author = "Red Team C2 Project"
        severity = "medium"

    strings:
        $py1 = "#!/usr/bin/env python" ascii
        $py2 = "import requests" ascii
        $py3 = "import click" ascii

        // C2 API calls
        $api1 = "/api/operator/beacons" ascii
        $api2 = "/api/operator/command" ascii
        $api3 = "/api/operator/outputs" ascii

        // CLI commands
        $cmd1 = "@cli.command()" ascii
        $cmd2 = "def beacons" ascii
        $cmd3 = "def exec" ascii

    condition:
        3 of ($py*) and 2 of ($api*)
}

// =====================================================================
// Rule 13: Detect Traffic Obfuscation
// =====================================================================

rule C2_Traffic_Obfuscation
{
    meta:
        description = "Detects ML-based traffic obfuscation"
        author = "Red Team C2 Project"
        severity = "medium"

    strings:
        // Machine learning libraries
        $ml1 = "tensorflow" ascii
        $ml2 = "import numpy" ascii
        $ml3 = "keras" ascii

        // GAN-related
        $gan1 = "Generator" ascii
        $gan2 = "Discriminator" ascii
        $gan3 = "GAN" ascii

        // Traffic generation
        $traffic1 = "traffic_gan" ascii
        $traffic2 = "generate_traffic" ascii
        $traffic3 = "PacketInterval" ascii

    condition:
        (1 of ($ml*)) and (1 of ($gan*))
}
