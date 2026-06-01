#!/bin/bash

# Script to manage Instagram related services

# Navigate to the script's directory to ensure relative paths are correct
cd "$(dirname "$0")/.." # Go to project root

# --- Configuration ---
LOG_DIR="./logs"
BINARY_DIR="./bin"
SERVICES=(
  "instagram_fetcher"
  "instagram_posts_parser"
  "instagram_clickhouse_sink"
  "instagram_immediate_processor"
)

# --- Functions ---

# Function to start all services
start_services() {
  echo "Starting Instagram services..."
  
  # Load environment variables from .env file
  if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs -d '\n')
  else
    echo "Error: .env file not found!"
    exit 1
  fi

  # Create logs directory if it doesn't exist
  mkdir -p "$LOG_DIR"

  for SERVICE_NAME in "${SERVICES[@]}"; do
    BINARY_PATH="$BINARY_DIR/$SERVICE_NAME"
    LOG_FILE="$LOG_DIR/${SERVICE_NAME}.log"

    if [ ! -f "$BINARY_PATH" ]; then
      echo "Error: Binary for $SERVICE_NAME not found at $BINARY_PATH. Please build the service."
      continue
    fi

    # Check if service is already running
    if pgrep -f "$BINARY_PATH" > /dev/null; then
      echo "$SERVICE_NAME is already running."
    else
      echo "Starting $SERVICE_NAME... Output will be logged to $LOG_FILE"
      nohup "$BINARY_PATH" > "$LOG_FILE" 2>&1 &
      sleep 1 # Give it a moment to start
      if pgrep -f "$BINARY_PATH" > /dev/null; then
        echo "$SERVICE_NAME started successfully."
      else
        echo "Error: $SERVICE_NAME failed to start. Check $LOG_FILE for details."
      fi
    fi
    echo "-------------------------------------"
  done
  echo "All Instagram services have been processed."
}

# Function to stop all services
stop_services() {
  echo "Stopping Instagram services..."
  for SERVICE_NAME in "${SERVICES[@]}"; do
    BINARY_PATH="$BINARY_DIR/$SERVICE_NAME"
    
    # Find process ID
    PID=$(pgrep -f "$BINARY_PATH")

    if [ -n "$PID" ]; then
      echo "Stopping $SERVICE_NAME (PID: $PID)..."
      kill $PID
      # Wait a bit and check if it's gone
      sleep 1
      if pgrep -f "$BINARY_PATH" > /dev/null; then
        echo "Failed to stop $SERVICE_NAME gracefully, sending SIGKILL..."
        kill -9 $PID
      else
        echo "$SERVICE_NAME stopped."
      fi
    else
      echo "$SERVICE_NAME is not running."
    fi
    echo "-------------------------------------"
  done
  echo "All Instagram services have been processed."
}

# --- Main Logic ---
ACTION=$1

case "$ACTION" in
  start)
    start_services
    ;;
  stop)
    stop_services
    ;;
  *)
    echo "Usage: $0 {start|stop}"
    exit 1
    ;;
esac
