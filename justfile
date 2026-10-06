COMPOSE_FILE := "deployments/local/docker-compose.yml"
COMPOSE_INFRA := "deployments/local/docker-compose.infra.yml"
SERVICES := "apigateway migrate auth role user category cashier merchant order_item order product transaction email stats_reader stats_writer"
LOCAL_SERVICES := "auth role user category cashier merchant order_item order product transaction apigateway"
DOCKER_COMPOSE := "docker compose"

migrate:
    go run service/migrate/main.go

seeder:
    go run service/seeder/main.go

generate-proto:
    cd proto && find . -name "*.proto" -not -path "./google/*" -not -path "./proto/*" -exec protoc --proto_path=. --go_out=../pb --go_opt=module=github.com/MamangRust/microservice-point-of-sale-pb --go-grpc_out=../pb --go-grpc_opt=module=github.com/MamangRust/microservice-point-of-sale-pb {} +

# Build all services that contain a go.mod file
build:
    @mkdir -p bin
    @for mod in service/*/go.mod; do \
        dir=$(dirname $mod); \
        service=$(basename $dir); \
        echo "🔨 Building $service..."; \
        if [ -f "$dir/cmd/main.go" ]; then \
            (cd $dir && go build -o ../../bin/$service ./cmd/main.go) || exit 1; \
        else \
            (cd $dir && go build -o ../../bin/$service ./main.go) || exit 1; \
        fi; \
    done
    @echo "✅ All services built successfully in bin/ folder."

generate-swagger:
    @command -v swag >/dev/null 2>&1 || (echo "swag tidak ada di PATH — install: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
    cd service/apigateway && swag init -g cmd/main.go -o docs --parseDependency --parseInternal
    @echo "Swagger docs regenerated: service/apigateway/docs"

build-image:
    @for service in $(SERVICES); do \
    	echo "🔨 Building $$service-pointofsale-service..."; \
    	docker build -t $$service-pointofsale-service:1.0 -f service/$$service/Dockerfile . || exit 1; \
    done
    @echo "✅ All services built successfully."

image-load:
    @for service in $(SERVICES); do \
    	echo "🚚 Loading $$service-pointofsale-service..."; \
    	minikube image load $$service-pointofsale-service:1.0 || exit 1; \
    done
    @echo "✅ All services loaded successfully."

ps:
    ${DOCKER_COMPOSE} -f $(COMPOSE_FILE) ps

up:
    ${DOCKER_COMPOSE} -f $(COMPOSE_FILE) up -d

down:
    ${DOCKER_COMPOSE} -f $(COMPOSE_FILE) down

# ── Infra-only (E2E lokal: service jalan LOKAL, compose hanya pg/redis/observability) ──
infra-up:
    ${DOCKER_COMPOSE} -f $(COMPOSE_INFRA) up -d

infra-down:
    ${DOCKER_COMPOSE} -f $(COMPOSE_INFRA) down

infra-ps:
    ${DOCKER_COMPOSE} -f $(COMPOSE_INFRA) ps

# Jalankan service lokal (tanpa kafka) — binary harus sudah di-build (just build)
services-up: build
    @mkdir -p logs
    @for s in $(LOCAL_SERVICES); do \
        echo "🚀 Start $$s..."; \
        KAFKA_BROKERS="" APP_ENV=development nohup ./bin/$$s >> logs/$$s.log 2>&1 & \
    done
    @echo "✅ Service lokal berjalan (log: logs/*.log). Email service tidak dijalankan (butuh kafka)."

services-down:
    @pkill -f 'bin/(auth|role|user|category|cashier|merchant|order_item|order|product|transaction|apigateway)' 2>/dev/null || true
    @echo "✅ Service lokal dihentikan."

# E2E lengkap: infra up → migrate → seed → service lokal → hurl semua endpoint
e2e-hurl:
    ./tests/hurl/run_e2e.sh

build-up:
    just build-image && just up

# F3 stats pipeline: materialize historical OLTP rows into ClickHouse once.
# Requires the full stack up (source services + kafka + clickhouse).
stats-backfill:
    ${DOCKER_COMPOSE} -f $(COMPOSE_FILE) run --rm stats_writer backfill

# Follow stats_reader/stats_writer logs.
stats-logs:
    ${DOCKER_COMPOSE} -f $(COMPOSE_FILE) logs -f stats_reader stats_writer

# Smoke test: verifikasi trace_id kontinu HTTP → gRPC → Kafka → email (via Jaeger)
smoke-trace:
    ./tests/smoke/trace_smoke.sh

# Fase 6: lifecycle infra compose — clean volume start, restart/stop/start/down-up tanpa kehilangan data
# ⚠ DESTRUKTIF: menjalankan `docker compose down -v` (menghapus volume infra).
# Pemakaian: just smoke-stack-lifecycle yes
smoke-stack-lifecycle approve:
    @if [ "{{approve}}" != "yes" ]; then echo "⚠ DESTRUKTIF: butuh konfirmasi. Jalankan: just smoke-stack-lifecycle yes"; exit 1; fi
    STACK_LIFECYCLE_CONFIRM=1 ./tests/smoke/stack_lifecycle.sh

# Tidy all go.mod files
tidy-all:
    @for mod in service/*/go.mod; do \
    	dir=$(dirname $mod); \
    	service=$(basename $dir); \
    	echo "🧹 Tidying $service..."; \
    	(cd $dir && go mod tidy) || exit 1; \
    done
    @echo "✅ All services tidied successfully."


# Run unit tests in pkg/
test-unit:
    @echo "🧪 Running unit tests in pkg/..."
    @cd pkg && go test ./... -v

# Run integration tests in tests/
test-integration:
    @echo "🧪 Running integration tests in tests/..."
    @cd tests && GOWORK=off APP_ENV=development go test ./... -v

# Run all tests (unit and integration)
test-all: test-unit test-integration
