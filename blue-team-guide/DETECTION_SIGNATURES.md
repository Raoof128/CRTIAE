# Red Team C2 Framework - Blue Team Detection Guide

## Executive Summary

This document provides comprehensive detection signatures and mitigation strategies for the Red Team C2 framework. As defenders, understanding attacker techniques is critical for effective detection and response.

**Target Audience**: Blue team analysts, SOC operators, incident responders, threat hunters

**Detection Coverage**:
- ✅ Network-level indicators (HTTPS beaconing, DNS tunneling)
- ✅ Host-level indicators (process behavior, file system artifacts)
- ✅ Database-level indicators (C2 infrastructure detection)
- ✅ Behavioral analytics (statistical beaconing analysis)

---

## Table of Contents

1. [Network-Level Detection](#network-level-detection)
2. [Host-Level Detection](#host-level-detection)
3. [Database-Level Detection](#database-level-detection)
4. [Behavioral Detection](#behavioral-detection)
5. [YARA Rules](#yara-rules)
6. [Suricata/Snort Rules](#suricatasnort-rules)
7. [Splunk Queries](#splunk-queries)
8. [Mitigation Strategies](#mitigation-strategies)

---

## Network-Level Detection

### 1.1 HTTPS Beaconing Detection

**Indicators of Compromise (IoCs)**:
- Periodic HTTPS connections (60-180 second intervals)
- Consistent connection to single external IP
- Unusual TLS cipher suites (ChaCha20-Poly1305 from desktop applications)
- High entropy POST payloads (encrypted C2 traffic)
- User-Agent inconsistencies (rotation pattern)

**Detection Methods**:

#### Zeek/Bro IDS Detection
```zeek
# Zeek script: detect-c2-beaconing.zeek
# Detects periodic HTTPS connections characteristic of C2 beaconing

@load base/protocols/http

module C2Detection;

export {
    redef enum Notice::Type += {
        C2_Beaconing_Detected
    };
}

global beacon_tracker: table[addr] of set[time];

event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
{
    local src = c$id$orig_h;

    # Track request timestamps per source IP
    if (src !in beacon_tracker)
        beacon_tracker[src] = set();

    add beacon_tracker[src][network_time()];

    # Check for periodic beaconing (interval analysis)
    if (|beacon_tracker[src]| >= 5) {
        # Statistical analysis of intervals
        # If standard deviation < 20% of mean, likely beaconing

        NOTICE([$note=C2_Beaconing_Detected,
                $msg=fmt("Possible C2 beaconing from %s", src),
                $src=src]);
    }
}
```

#### Splunk Detection Query
```spl
index=proxy OR index=firewall
| stats count by src_ip dest_ip dest_port
| where dest_port=443 AND count > 10
| eventstats avg(eval(round(_time,0))) as avg_interval by src_ip dest_ip
| eventstats stdev(eval(round(_time,0))) as stdev_interval by src_ip dest_ip
| eval cv = stdev_interval / avg_interval
| where cv < 0.2
| table src_ip dest_ip count avg_interval cv
| rename cv as "Coefficient_of_Variation"
```

**Explanation**: Low coefficient of variation (<0.2) indicates periodic beaconing.

---

### 1.2 DNS Covert Channel Detection

**Indicators**:
- Unusually long subdomain queries (>50 characters)
- High volume of TXT record queries to single domain
- Base32 encoding patterns in DNS queries
- DNS query frequency exceeds baseline (>100 queries/minute)

**Suricata Rule**:
```suricata
# Detect DNS tunneling via subdomain length
alert dns any any -> any 53 (msg:"DNS Tunneling - Long Subdomain";
    dns_query; content:".";
    pcre:"/^[a-z0-9]{50,}\./i";
    threshold:type limit, track by_src, count 5, seconds 60;
    classtype:trojan-activity;
    sid:1000001; rev:1;)

# Detect excessive TXT queries
alert dns any any -> any 53 (msg:"DNS Tunneling - Excessive TXT Queries";
    dns_query; dns.query.type:16;
    threshold:type threshold, track by_src, count 10, seconds 60;
    classtype:trojan-activity;
    sid:1000002; rev:1;)

# Detect Base32 encoding in DNS
alert dns any any -> any 53 (msg:"DNS Tunneling - Base32 Pattern";
    dns_query;
    pcre:"/[A-Z2-7]{20,}/";
    classtype:trojan-activity;
    sid:1000003; rev:1;)
```

**Python Detection Script**:
```python
#!/usr/bin/env python3
"""
DNS Tunneling Detector
Analyzes DNS logs for covert channel indicators
"""

import re
from collections import defaultdict
from datetime import datetime

def detect_dns_tunneling(dns_logs):
    """
    Analyzes DNS query logs for tunneling patterns

    Returns: List of suspicious domains
    """
    domain_stats = defaultdict(lambda: {
        'count': 0,
        'avg_subdomain_len': 0,
        'txt_queries': 0
    })

    for log in dns_logs:
        domain = log['query_domain']
        subdomain = domain.split('.')[0]

        domain_stats[domain]['count'] += 1
        domain_stats[domain]['avg_subdomain_len'] += len(subdomain)

        if log['query_type'] == 'TXT':
            domain_stats[domain]['txt_queries'] += 1

    # Identify suspicious domains
    suspicious = []
    for domain, stats in domain_stats.items():
        avg_len = stats['avg_subdomain_len'] / stats['count']

        # Flags:
        # - Average subdomain length > 40
        # - More than 10 TXT queries
        # - Query frequency > 50/minute
        if avg_len > 40 or stats['txt_queries'] > 10 or stats['count'] > 50:
            suspicious.append({
                'domain': domain,
                'avg_subdomain_length': avg_len,
                'txt_queries': stats['txt_queries'],
                'total_queries': stats['count']
            })

    return suspicious
```

---

## Host-Level Detection

### 2.1 Process Behavior Analysis

**Indicators**:
- Unsigned binary executing network connections
- Suspicious parent-child process relationships
- Command-line execution (cmd.exe, powershell.exe) from unusual parents
- Memory contains ChaCha20 constants

**Sysmon Configuration**:
```xml
<!-- Sysmon config for C2 detection -->
<Sysmon schemaversion="4.50">
  <EventFiltering>
    <!-- Process Creation -->
    <ProcessCreate onmatch="include">
      <!-- Detect unsigned binaries with network activity -->
      <Rule groupRelation="and">
        <Image condition="not contains">C:\Windows\</Image>
        <Image condition="not contains">C:\Program Files\</Image>
        <Signed condition="is">false</Signed>
      </Rule>

      <!-- Suspicious PowerShell executions -->
      <CommandLine condition="contains">-ExecutionPolicy Bypass</CommandLine>
      <CommandLine condition="contains">-enc</CommandLine>
      <CommandLine condition="contains">-nop</CommandLine>
    </ProcessCreate>

    <!-- Network Connections -->
    <NetworkConnect onmatch="include">
      <Image condition="not contains">C:\Windows\</Image>
      <Initiated condition="is">true</Initiated>
      <DestinationPort condition="is">443</DestinationPort>
    </NetworkConnect>

    <!-- Image Load (DLL monitoring) -->
    <ImageLoad onmatch="include">
      <Image condition="contains">winhttp.dll</Image>
      <Image condition="contains">bcrypt.dll</Image>
    </ImageLoad>
  </EventFiltering>
</Sysmon>
```

**Windows Event Log Detection**:

| Event ID | Description | Detection Logic |
|----------|-------------|-----------------|
| 4688 | Process Creation | Monitor for unsigned .exe from TEMP/Downloads spawning cmd.exe |
| 3 (Sysmon) | Network Connection | Unsigned process connecting to external IP:443 |
| 7 (Sysmon) | Image Load | Loading of WinHTTP, BCrypt, or crypto libraries |
| 13 (Sysmon) | Registry Set | AMSI bypass attempts (AmsiEnable=0) |
| 4104 | PowerShell ScriptBlock | Commands with `-enc`, `-ExecutionPolicy Bypass` |

**Splunk Query for Process Detection**:
```spl
index=windows EventCode=4688 OR EventCode=3
| where (EventCode=4688 AND (like(NewProcessName, "%\\Temp\\%") OR like(NewProcessName, "%\\Downloads\\%")))
| join ProcessId [search EventCode=3 DestinationPort=443]
| stats count by Computer NewProcessName DestinationIp
| where count > 5
```

---

### 2.2 Memory Analysis

**Volatility Plugin** (Custom):
```python
# volatility_c2_detector.py
# Volatility plugin to detect C2 implants in memory

import volatility.plugins.common as common
import volatility.utils as utils
import re

class C2Detector(common.AbstractWindowsCommand):
    """Detects C2 implant signatures in process memory"""

    def calculate(self):
        addr_space = utils.load_as(self._config)

        # ChaCha20 constants
        chacha_constants = [
            b"expand 32-byte k",
            b"expa",  # Partial match
        ]

        # Suspicious strings
        suspicious_strings = [
            b"/api/beacon/",
            b"/api/commands/",
            b"BeaconID",
            b"ChaCha20",
        ]

        for proc in tasks.pslist(addr_space):
            proc_as = proc.get_process_address_space()
            if not proc_as:
                continue

            # Scan process memory
            for const in chacha_constants + suspicious_strings:
                for hit in proc_as.zread(proc.Peb.ImageBaseAddress, 0x10000):
                    if const in hit:
                        yield proc.ImageFileName, proc.UniqueProcessId, const
```

---

## Database-Level Detection

### 3.1 PostgreSQL C2 Infrastructure Detection

**Indicators**:
- Database tables named: `beacons`, `commands`, `outputs`, `implants`
- High volume of INSERT operations (beacon check-ins)
- Encrypted TEXT columns (Base64 patterns)

**PostgreSQL Audit Query**:
```sql
-- Detect C2-like database schemas
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE tablename IN ('beacons', 'commands', 'outputs', 'implants', 'sessions')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Detect high insert frequency (beaconing)
SELECT
    relname AS table_name,
    n_tup_ins AS inserts,
    n_tup_upd AS updates
FROM pg_stat_user_tables
WHERE n_tup_ins > 1000
ORDER BY n_tup_ins DESC;
```

---

## Behavioral Detection

### 4.1 Statistical Beaconing Analysis

**Algorithm**: Detect periodic network connections using Fourier Transform

```python
import numpy as np
from scipy import signal
from datetime import datetime

def detect_beaconing_fft(timestamps, threshold=0.7):
    """
    Uses FFT to detect periodic beaconing

    Args:
        timestamps: List of connection timestamps
        threshold: Correlation threshold (0-1)

    Returns:
        True if periodic beaconing detected
    """
    if len(timestamps) < 10:
        return False

    # Convert to time deltas
    deltas = np.diff(sorted(timestamps))

    # Apply FFT
    fft = np.fft.fft(deltas)
    power = np.abs(fft) ** 2

    # Find dominant frequency
    dominant_freq = np.argmax(power[1:len(power)//2]) + 1

    # Correlation test
    correlation = power[dominant_freq] / np.sum(power)

    return correlation > threshold
```

---

## YARA Rules

Complete YARA rules in `/blue-team-guide/yara_rules.yara`

Key rules:
- `C2_HTTPS_Beacon_Implant`
- `C2_ChaCha20_Crypto`
- `C2_Teamserver_Rust`
- `C2_DNS_Tunneling`

See full rules in dedicated YARA file.

---

## Mitigation Strategies

### 5.1 Network Mitigations

1. **TLS Inspection**: Deploy SSL/TLS inspection at proxy level
2. **DNS Filtering**: Block excessive subdomain queries, rate-limit TXT queries
3. **Geo-blocking**: Restrict HTTPS connections to approved countries
4. **Threat Intelligence**: Subscribe to C2 domain feeds (abuse.ch, Emerging Threats)

### 5.2 Host Mitigations

1. **Application Whitelisting**: Only allow signed binaries to execute
2. **EDR Deployment**: Behavioral analysis engines (CrowdStrike, SentinelOne, Defender ATP)
3. **PowerShell Hardening**:
   - Enable ScriptBlock logging (Event 4104)
   - Enable Module logging
   - Constrained Language Mode
4. **AMSI Protection**: Lock AMSI registry keys, prevent tamper

### 5.3 Defense-in-Depth

```
┌─────────────────────────────────────────────────┐
│  Layer 1: Network Perimeter                    │
│  - Firewall rules (block unusual ports)        │
│  - IDS/IPS (Suricata rules)                    │
│  - DNS filtering                               │
├─────────────────────────────────────────────────┤
│  Layer 2: Endpoint Protection                  │
│  - EDR (behavioral monitoring)                 │
│  - Application whitelisting                    │
│  - AMSI/ETW protection                         │
├─────────────────────────────────────────────────┤
│  Layer 3: Identity & Access                    │
│  - MFA for privileged accounts                 │
│  - Least privilege                             │
│  - Credential Guard (Windows)                  │
├─────────────────────────────────────────────────┤
│  Layer 4: Monitoring & Response                │
│  - SIEM (Splunk, ELK)                          │
│  - Threat hunting (regular sweeps)             │
│  - Incident response playbooks                 │
└─────────────────────────────────────────────────┘
```

---

## Incident Response Playbook

**Phase 1: Detection**
1. Alert triggered (beaconing detected via Zeek/Splunk)
2. Analyst validates alert (reduce false positives)

**Phase 2: Containment**
1. Isolate affected host from network
2. Block C2 domain/IP at firewall
3. Disable compromised user account

**Phase 3: Eradication**
1. Identify implant process (Sysmon, EDR)
2. Kill malicious process
3. Delete implant binary
4. Clear persistence mechanisms (registry, scheduled tasks)

**Phase 4: Recovery**
1. Reimage affected host (recommended)
2. Restore from clean backup
3. Reset all credentials
4. Monitor for re-infection (30 days)

**Phase 5: Lessons Learned**
1. Root cause analysis
2. Update detection rules
3. Improve security controls

---

## Detection Maturity Model

| Level | Capability | Tools Required |
|-------|------------|----------------|
| **1 - Basic** | Signature-based AV | Windows Defender |
| **2 - Intermediate** | Network monitoring | Zeek, Suricata |
| **3 - Advanced** | Behavioral analytics | EDR, SIEM |
| **4 - Expert** | Threat hunting | Custom scripts, ML models |
| **5 - Optimized** | Automated response | SOAR platform |

**Recommendation**: Aim for Level 3+ to detect sophisticated C2 frameworks

---

## Summary

This C2 framework provides **dual value**:
- **Red Team**: Understanding offensive capabilities
- **Blue Team**: Building detection and response strategies

**Key Takeaway**: Modern C2 frameworks use encryption, jitter, and protocol blending. Detection requires multi-layered approach combining network analytics, endpoint telemetry, and behavioral analysis.

**For Defenders**: Focus on behavioral detection over signature-based. Attackers will always modify malware signatures, but beaconing behavior is fundamental to C2 operation.

---

**Document Version**: 1.0
**Last Updated**: 2024
**Author**: Red Team C2 Project - Macquarie University
**Contact**: See repository for responsible disclosure policy
