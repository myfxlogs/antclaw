#!/bin/bash
# Production deployment script for AntClaw
set -e

echo "=== AntClaw Production Deployment ==="

# Step 1: Build and start containers
echo "Building and starting containers..."
docker compose -f deploy/docker-compose.yaml up -d --build

# Step 2: Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
for i in {1..30}; do
    if docker compose -f deploy/docker-compose.yaml exec -T postgres pg_isready -U antclaw 2>/dev/null; then
        echo "PostgreSQL is ready!"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 2
done

# Step 3: Run database migrations
echo "Running database migrations..."
for file in /opt/antclaw/backend/db/migrations/*.sql; do
    echo "Applying: $(basename "$file")"
    docker compose -f deploy/docker-compose.yaml exec -T postgres psql -U antclaw -d antclaw -f - < "$file" 2>&1 | tail -3
done

# Step 4: Verify deployment
echo "Verifying deployment..."
docker compose -f deploy/docker-compose.yaml ps

echo ""
echo "=== Deployment Complete ==="
echo "Services:"
echo "- Web: http://localhost:8080"
echo "- Admin: http://localhost:8081"
echo "- API: http://localhost:8082"
echo "- MinIO Console: http://localhost:9001"
echo "- PostgreSQL: localhost:5433"
echo "- Redis: localhost:6379"

# Show table count
echo ""
echo "Database tables:"
docker compose -f deploy/docker-compose.yaml exec -T postgres psql -U antclaw -d antclaw -tc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null || echo "Could not query table count"
