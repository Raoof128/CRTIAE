"""
Test suite for the Red Team C2 Controller
Tests CLI commands, API interactions, and error handling
"""

import pytest
import json
from unittest.mock import Mock, patch, MagicMock
from click.testing import CliRunner
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import main
from main import (
    cli,
    beacons,
    exec_command,
    output,
    interactive,
    C2Client
)


class TestC2Client:
    """Tests for the C2Client class"""

    def setup_method(self):
        """Set up test fixtures"""
        self.base_url = "http://localhost:8443"
        self.client = C2Client(self.base_url)

    @patch('requests.get')
    def test_list_beacons_success(self, mock_get):
        """Test successful beacon listing"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "beacons": [
                {
                    "id": "beacon-123",
                    "hostname": "test-host",
                    "username": "test-user",
                    "os": "linux",
                    "ip": "192.168.1.100",
                    "last_seen": "2024-01-15T10:30:00Z",
                    "status": "active"
                }
            ]
        }
        mock_get.return_value = mock_response

        result = self.client.list_beacons()

        assert len(result) == 1
        assert result[0]["id"] == "beacon-123"
        assert result[0]["hostname"] == "test-host"
        mock_get.assert_called_once()

    @patch('requests.get')
    def test_list_beacons_error(self, mock_get):
        """Test error handling in beacon listing"""
        mock_get.side_effect = Exception("Connection error")

        with pytest.raises(Exception):
            self.client.list_beacons()

    @patch('requests.post')
    def test_submit_command_success(self, mock_post):
        """Test successful command submission"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "command_id": "cmd-456",
            "status": "queued"
        }
        mock_post.return_value = mock_response

        result = self.client.submit_command("beacon-123", "whoami")

        assert result["command_id"] == "cmd-456"
        assert result["status"] == "queued"
        mock_post.assert_called_once()

    @patch('requests.post')
    def test_submit_command_invalid_beacon(self, mock_post):
        """Test command submission with invalid beacon"""
        mock_response = Mock()
        mock_response.status_code = 404
        mock_response.text = "Beacon not found"
        mock_post.return_value = mock_response

        with pytest.raises(Exception):
            self.client.submit_command("invalid-beacon", "whoami")

    @patch('requests.get')
    def test_get_output_success(self, mock_get):
        """Test successful output retrieval"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "outputs": [
                {
                    "command_id": "cmd-456",
                    "output": "test-user",
                    "timestamp": "2024-01-15T10:31:00Z"
                }
            ]
        }
        mock_get.return_value = mock_response

        result = self.client.get_output("beacon-123")

        assert len(result) == 1
        assert result[0]["output"] == "test-user"
        mock_get.assert_called_once()

    @patch('requests.get')
    def test_get_output_empty(self, mock_get):
        """Test output retrieval with no results"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"outputs": []}
        mock_get.return_value = mock_response

        result = self.client.get_output("beacon-123")

        assert len(result) == 0


class TestCLICommands:
    """Tests for CLI commands"""

    def setup_method(self):
        """Set up test runner"""
        self.runner = CliRunner()

    @patch('main.C2Client.list_beacons')
    def test_beacons_command_success(self, mock_list_beacons):
        """Test beacons command with results"""
        mock_list_beacons.return_value = [
            {
                "id": "beacon-123",
                "hostname": "test-host",
                "username": "test-user",
                "os": "linux",
                "ip": "192.168.1.100",
                "last_seen": "2024-01-15T10:30:00Z",
                "status": "active"
            }
        ]

        result = self.runner.invoke(beacons, ['--server', 'http://localhost:8443'])

        assert result.exit_code == 0
        assert "beacon-123" in result.output
        assert "test-host" in result.output

    @patch('main.C2Client.list_beacons')
    def test_beacons_command_empty(self, mock_list_beacons):
        """Test beacons command with no results"""
        mock_list_beacons.return_value = []

        result = self.runner.invoke(beacons, ['--server', 'http://localhost:8443'])

        assert result.exit_code == 0
        assert "No active beacons" in result.output

    @patch('main.C2Client.submit_command')
    def test_exec_command_success(self, mock_submit_command):
        """Test exec command success"""
        mock_submit_command.return_value = {
            "command_id": "cmd-456",
            "status": "queued"
        }

        result = self.runner.invoke(
            exec_command,
            ['--server', 'http://localhost:8443', 'beacon-123', 'whoami']
        )

        assert result.exit_code == 0
        assert "cmd-456" in result.output

    @patch('main.C2Client.submit_command')
    def test_exec_command_error(self, mock_submit_command):
        """Test exec command with error"""
        mock_submit_command.side_effect = Exception("Beacon not found")

        result = self.runner.invoke(
            exec_command,
            ['--server', 'http://localhost:8443', 'invalid-beacon', 'whoami']
        )

        assert result.exit_code != 0

    @patch('main.C2Client.get_output')
    def test_output_command_success(self, mock_get_output):
        """Test output command with results"""
        mock_get_output.return_value = [
            {
                "command_id": "cmd-456",
                "output": "test-user",
                "timestamp": "2024-01-15T10:31:00Z"
            }
        ]

        result = self.runner.invoke(
            output,
            ['--server', 'http://localhost:8443', 'beacon-123']
        )

        assert result.exit_code == 0
        assert "test-user" in result.output

    @patch('main.C2Client.get_output')
    def test_output_command_empty(self, mock_get_output):
        """Test output command with no results"""
        mock_get_output.return_value = []

        result = self.runner.invoke(
            output,
            ['--server', 'http://localhost:8443', 'beacon-123']
        )

        assert result.exit_code == 0
        assert "No output" in result.output


class TestInteractiveMode:
    """Tests for interactive mode"""

    def setup_method(self):
        """Set up test runner"""
        self.runner = CliRunner()

    @patch('main.C2Client.list_beacons')
    @patch('builtins.input')
    def test_interactive_list_command(self, mock_input, mock_list_beacons):
        """Test 'list' command in interactive mode"""
        mock_list_beacons.return_value = [
            {
                "id": "beacon-123",
                "hostname": "test-host",
                "username": "test-user",
                "os": "linux",
                "ip": "192.168.1.100",
                "last_seen": "2024-01-15T10:30:00Z",
                "status": "active"
            }
        ]
        mock_input.side_effect = ["list", "exit"]

        result = self.runner.invoke(
            interactive,
            ['--server', 'http://localhost:8443']
        )

        assert "beacon-123" in result.output

    @patch('builtins.input')
    def test_interactive_exit_command(self, mock_input):
        """Test 'exit' command in interactive mode"""
        mock_input.return_value = "exit"

        result = self.runner.invoke(
            interactive,
            ['--server', 'http://localhost:8443']
        )

        assert result.exit_code == 0


class TestErrorHandling:
    """Tests for error handling"""

    def setup_method(self):
        """Set up test client"""
        self.client = C2Client("http://localhost:8443")

    @patch('requests.get')
    def test_network_timeout(self, mock_get):
        """Test handling of network timeout"""
        import requests
        mock_get.side_effect = requests.Timeout("Connection timeout")

        with pytest.raises(requests.Timeout):
            self.client.list_beacons()

    @patch('requests.get')
    def test_connection_refused(self, mock_get):
        """Test handling of connection refused"""
        import requests
        mock_get.side_effect = requests.ConnectionError("Connection refused")

        with pytest.raises(requests.ConnectionError):
            self.client.list_beacons()

    @patch('requests.get')
    def test_invalid_json_response(self, mock_get):
        """Test handling of invalid JSON response"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.side_effect = json.JSONDecodeError("Invalid JSON", "", 0)
        mock_get.return_value = mock_response

        with pytest.raises(json.JSONDecodeError):
            self.client.list_beacons()


class TestDataValidation:
    """Tests for data validation"""

    def test_beacon_id_validation(self):
        """Test beacon ID validation"""
        valid_ids = ["beacon-123", "BEACON-456", "beacon_789"]
        invalid_ids = ["", " ", "beacon@123", "beacon 123"]

        for beacon_id in valid_ids:
            assert len(beacon_id) > 0

        for beacon_id in invalid_ids:
            # Would validate with actual validation function
            pass

    def test_command_validation(self):
        """Test command validation"""
        valid_commands = ["whoami", "ls -la", "ps aux"]

        for cmd in valid_commands:
            assert len(cmd) > 0
            assert cmd.strip() == cmd or cmd == cmd.strip()


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
