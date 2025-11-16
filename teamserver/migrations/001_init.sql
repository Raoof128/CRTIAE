-- Initial database schema for C2 teamserver
--
-- BLUE TEAM DETECTION:
-- Monitor for database creation with suspicious table names:
-- - "beacons", "commands", "outputs", "implants"
-- - Audit PostgreSQL logs for C2-like schema patterns

-- Beacons table - Active implants
CREATE TABLE IF NOT EXISTS beacons (
    id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    username VARCHAR(255),
    os VARCHAR(100),
    ip VARCHAR(50),
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Indexes for performance
    CONSTRAINT unique_beacon_id UNIQUE (id)
);

CREATE INDEX idx_beacons_last_seen ON beacons(last_seen DESC);
CREATE INDEX idx_beacons_status ON beacons(status);

-- Commands table - Operator-issued commands
CREATE TABLE IF NOT EXISTS commands (
    id VARCHAR(255) PRIMARY KEY,
    beacon_id VARCHAR(255) NOT NULL,
    command_type VARCHAR(100) NOT NULL,
    command_data TEXT NOT NULL, -- Encrypted command payload
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    executed BOOLEAN NOT NULL DEFAULT false,
    executed_at TIMESTAMP,

    -- Foreign key
    CONSTRAINT fk_beacon
        FOREIGN KEY (beacon_id)
        REFERENCES beacons(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_commands_beacon_id ON commands(beacon_id);
CREATE INDEX idx_commands_executed ON commands(executed);
CREATE INDEX idx_commands_created_at ON commands(created_at DESC);

-- Outputs table - Command execution results
CREATE TABLE IF NOT EXISTS outputs (
    id VARCHAR(255) PRIMARY KEY,
    beacon_id VARCHAR(255) NOT NULL,
    command_id VARCHAR(255),
    output_data TEXT NOT NULL, -- Encrypted output
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'success',

    -- Foreign keys
    CONSTRAINT fk_output_beacon
        FOREIGN KEY (beacon_id)
        REFERENCES beacons(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_output_command
        FOREIGN KEY (command_id)
        REFERENCES commands(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_outputs_beacon_id ON outputs(beacon_id);
CREATE INDEX idx_outputs_timestamp ON outputs(timestamp DESC);
CREATE INDEX idx_outputs_command_id ON outputs(command_id);

-- Operations log - Audit trail
CREATE TABLE IF NOT EXISTS operations_log (
    id SERIAL PRIMARY KEY,
    operator VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    target_beacon VARCHAR(255),
    details TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_operations_log_timestamp ON operations_log(timestamp DESC);
CREATE INDEX idx_operations_log_operator ON operations_log(operator);
