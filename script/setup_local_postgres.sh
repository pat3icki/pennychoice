#!/bin/bash
# setup_local_postgres.sh – Start a PostgreSQL container (Alpine) for local development,
# then generate local_db.env with the connection URL.

set -e

# Default values
CONTAINER_NAME="postgres-dev"
POSTGRES_USER="myapp"
POSTGRES_PASSWORD="secret"
POSTGRES_DB="myapp"
PORT_MAP="5432:5432"
IMAGE="postgres:alpine"
RUN_QUERY=""   # optional SQL to execute after startup

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --query)
            RUN_QUERY="$2"
            shift 2
            ;;
        --name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        --password)
            POSTGRES_PASSWORD="$2"
            shift 2
            ;;
        --user)
            POSTGRES_USER="$2"
            shift 2
            ;;
        --db)
            POSTGRES_DB="$2"
            shift 2
            ;;
        --port)
            PORT_MAP="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --query 'SQL'    Run a SQL query after the container is ready"
            echo "  --name NAME      Container name (default: postgres-dev)"
            echo "  --password PASS  PostgreSQL password (default: secret)"
            echo "  --user USER      PostgreSQL user (default: myapp)"
            echo "  --db DB          Database name (default: myapp)"
            echo "  --port MAP       Port mapping (default: 5432:5432)"
            echo "  --help           Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker not installed."
    exit 1
fi

# Remove existing container if present
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container ${CONTAINER_NAME} exists. Stopping and removing..."
    docker stop "${CONTAINER_NAME}" &> /dev/null || true
    docker rm "${CONTAINER_NAME}" &> /dev/null || true
fi

# Pull the image (quietly)
echo "Pulling ${IMAGE}..."
docker pull -q "${IMAGE}"

# Start container
echo "Starting PostgreSQL container..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
    -e POSTGRES_USER="${POSTGRES_USER}" \
    -e POSTGRES_DB="${POSTGRES_DB}" \
    -p "${PORT_MAP}" \
    "${IMAGE}" > /dev/null

# Wait for PostgreSQL to be ready
echo -n "Waiting for PostgreSQL to accept connections"
for i in {1..30}; do
    if docker exec "${CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" &> /dev/null; then
        echo " ✅"
        break
    fi
    echo -n "."
    sleep 1
done

# If still not ready, exit
if ! docker exec "${CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" &> /dev/null; then
    echo "Error: PostgreSQL did not start in time."
    docker logs --tail 20 "${CONTAINER_NAME}"
    exit 1
fi

# Print connection info
echo "PostgreSQL is running:"
echo "  Host:      localhost"
echo "  Port:      ${PORT_MAP%%:*}"
echo "  User:      ${POSTGRES_USER}"
echo "  Password:  ${POSTGRES_PASSWORD}"
echo "  Database:  ${POSTGRES_DB}"
echo "Connect with: psql -h localhost -p ${PORT_MAP%%:*} -U ${POSTGRES_USER} -d ${POSTGRES_DB}"

# Run a query if provided
if [[ -n "${RUN_QUERY}" ]]; then
    echo "Executing query: ${RUN_QUERY}"
    docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" "${CONTAINER_NAME}" \
        psql -h localhost -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "${RUN_QUERY}"
fi

# -------------------------------------------------------------------
# Generate local_db.env using the Go program
# -------------------------------------------------------------------
echo "Generating local_db.env with the connection URL..."
go run ./script/gen_local_env.go \
    -POSTGRES_USER="${POSTGRES_USER}" \
    -POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
    -POSTGRES_DB="${POSTGRES_DB}" \
    -PORT_MAP="${PORT_MAP}"