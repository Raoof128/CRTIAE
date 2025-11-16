#!/usr/bin/env python3
"""
Red Team C2 Framework - Python API Client Library

Full-featured Python client for programmatic access to the C2 infrastructure.
Provides object-oriented interface to all teamserver endpoints.

Usage:
    from python_client import C2Client

    client = C2Client("https://teamserver:8443")
    beacons = client.beacons.list()
    client.commands.submit(beacon_id, "whoami")
"""

import requests
import json
import time
from typing import List, Dict, Optional, Any
from dataclasses import dataclass
from datetime import datetime
import urllib3

# Disable SSL warnings for testing
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


@dataclass
class Beacon:
    """Represents a C2 beacon/implant"""
    id: str
    hostname: str
    username: str
    os: str
    ip: str
    first_seen: str
    last_seen: str
    status: str

    @classmethod
    def from_dict(cls, data: Dict) -> 'Beacon':
        """Create Beacon from API response dict"""
        return cls(
            id=data.get('id', ''),
            hostname=data.get('hostname', ''),
            username=data.get('username', ''),
            os=data.get('os', ''),
            ip=data.get('ip', ''),
            first_seen=data.get('first_seen', ''),
            last_seen=data.get('last_seen', ''),
            status=data.get('status', 'unknown')
        )

    def __str__(self) -> str:
        return f"Beacon({self.id}, {self.hostname}, {self.username}@{self.os})"


@dataclass
class Command:
    """Represents a queued command"""
    id: str
    beacon_id: str
    command: str
    timestamp: str
    status: str

    @classmethod
    def from_dict(cls, data: Dict) -> 'Command':
        """Create Command from API response dict"""
        return cls(
            id=data.get('id', ''),
            beacon_id=data.get('beacon_id', ''),
            command=data.get('command', ''),
            timestamp=data.get('timestamp', ''),
            status=data.get('status', 'unknown')
        )


@dataclass
class Output:
    """Represents command output"""
    command_id: str
    beacon_id: str
    output: str
    timestamp: str

    @classmethod
    def from_dict(cls, data: Dict) -> 'Output':
        """Create Output from API response dict"""
        return cls(
            command_id=data.get('command_id', ''),
            beacon_id=data.get('beacon_id', ''),
            output=data.get('output', ''),
            timestamp=data.get('timestamp', '')
        )


class BeaconAPI:
    """API client for beacon operations"""

    def __init__(self, base_url: str, session: requests.Session):
        self.base_url = base_url
        self.session = session

    def list(self) -> List[Beacon]:
        """List all active beacons"""
        response = self.session.get(f"{self.base_url}/api/beacons")
        response.raise_for_status()
        data = response.json()
        return [Beacon.from_dict(b) for b in data.get('beacons', [])]

    def get(self, beacon_id: str) -> Optional[Beacon]:
        """Get specific beacon by ID"""
        beacons = self.list()
        for beacon in beacons:
            if beacon.id == beacon_id:
                return beacon
        return None

    def active(self, minutes: int = 5) -> List[Beacon]:
        """Get beacons active within last N minutes"""
        all_beacons = self.list()
        now = datetime.now()
        active_beacons = []

        for beacon in all_beacons:
            try:
                last_seen = datetime.fromisoformat(beacon.last_seen.replace('Z', '+00:00'))
                diff = (now - last_seen).total_seconds() / 60
                if diff <= minutes:
                    active_beacons.append(beacon)
            except:
                pass

        return active_beacons


class CommandAPI:
    """API client for command operations"""

    def __init__(self, base_url: str, session: requests.Session):
        self.base_url = base_url
        self.session = session

    def submit(self, beacon_id: str, command: str) -> Dict[str, Any]:
        """Submit a command to a beacon"""
        payload = {
            "beacon_id": beacon_id,
            "command": command
        }
        response = self.session.post(
            f"{self.base_url}/api/commands",
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def get_output(self, beacon_id: str, command_id: Optional[str] = None) -> List[Output]:
        """Get command output for a beacon"""
        url = f"{self.base_url}/api/output/{beacon_id}"
        if command_id:
            url += f"?command_id={command_id}"

        response = self.session.get(url)
        response.raise_for_status()
        data = response.json()
        return [Output.from_dict(o) for o in data.get('outputs', [])]

    def execute_and_wait(self, beacon_id: str, command: str, timeout: int = 300) -> str:
        """
        Execute command and wait for output

        Args:
            beacon_id: Target beacon ID
            command: Command to execute
            timeout: Max wait time in seconds

        Returns:
            Command output as string

        Raises:
            TimeoutError if output not received within timeout
        """
        # Submit command
        result = self.submit(beacon_id, command)
        command_id = result.get('command_id')

        if not command_id:
            raise ValueError("No command_id in response")

        # Poll for output
        start_time = time.time()
        while time.time() - start_time < timeout:
            outputs = self.get_output(beacon_id, command_id)
            if outputs:
                return outputs[0].output

            time.sleep(5)  # Poll every 5 seconds

        raise TimeoutError(f"Command output not received within {timeout} seconds")


class C2Client:
    """
    Main C2 API client

    Example usage:
        client = C2Client("https://teamserver:8443")

        # List beacons
        beacons = client.beacons.list()

        # Submit command
        result = client.commands.submit(beacon_id, "whoami")

        # Get output
        outputs = client.commands.get_output(beacon_id)
    """

    def __init__(self, base_url: str, verify_ssl: bool = False, timeout: int = 30):
        """
        Initialize C2 client

        Args:
            base_url: Teamserver URL (e.g., "https://teamserver:8443")
            verify_ssl: Whether to verify SSL certificates
            timeout: Request timeout in seconds
        """
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.verify = verify_ssl
        self.session.timeout = timeout

        # Initialize API modules
        self.beacons = BeaconAPI(self.base_url, self.session)
        self.commands = CommandAPI(self.base_url, self.session)

    def health_check(self) -> bool:
        """Check if teamserver is healthy"""
        try:
            response = self.session.get(f"{self.base_url}/health", timeout=5)
            return response.status_code == 200
        except:
            return False


# Example usage
if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="C2 Python API Client")
    parser.add_argument("--server", required=True, help="Teamserver URL")
    parser.add_argument("action", choices=["list", "exec", "output", "health"])
    parser.add_argument("--beacon", help="Beacon ID")
    parser.add_argument("--command", help="Command to execute")
    parser.add_argument("--wait", action="store_true", help="Wait for output")

    args = parser.parse_args()

    # Create client
    client = C2Client(args.server)

    if args.action == "health":
        if client.health_check():
            print("✓ Teamserver is healthy")
        else:
            print("✗ Teamserver is unhealthy")

    elif args.action == "list":
        beacons = client.beacons.list()
        print(f"Active Beacons: {len(beacons)}")
        for beacon in beacons:
            print(f"  {beacon}")

    elif args.action == "exec":
        if not args.beacon or not args.command:
            print("Error: --beacon and --command required")
            exit(1)

        if args.wait:
            # Execute and wait for output
            print(f"Executing: {args.command}")
            try:
                output = client.commands.execute_and_wait(args.beacon, args.command)
                print(f"Output:\n{output}")
            except TimeoutError as e:
                print(f"Error: {e}")
        else:
            # Just submit command
            result = client.commands.submit(args.beacon, args.command)
            print(f"Command submitted: {result.get('command_id')}")

    elif args.action == "output":
        if not args.beacon:
            print("Error: --beacon required")
            exit(1)

        outputs = client.commands.get_output(args.beacon)
        print(f"Outputs: {len(outputs)}")
        for output in outputs:
            print(f"\n[{output.timestamp}] Command: {output.command_id}")
            print(output.output)
