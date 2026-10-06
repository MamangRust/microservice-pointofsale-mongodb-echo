FROM golang:1.25.7-bookworm AS builder

WORKDIR /app

# Copy workspace and dependency manifests
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY shared/go.mod shared/go.sum* ./shared/
COPY pb/go.mod pb/go.sum* ./pb/
COPY tests/go.mod tests/go.sum* ./tests/

# Copy all service go.mod files
COPY service/ ./service/
RUN find service -type f ! -name "go.mod" ! -name "go.sum" -delete

# F1: per-service migrations. The find-delete above strips the SQL files, so
# re-copy each service's database/migration dir into the builder image to
# preserve them for the runtime stage.
COPY service/auth/database/migration/ ./service/auth/database/migration/
COPY service/cashier/database/migration/ ./service/cashier/database/migration/
COPY service/category/database/migration/ ./service/category/database/migration/
COPY service/email/database/migration/ ./service/email/database/migration/
COPY service/merchant/database/migration/ ./service/merchant/database/migration/
COPY service/order/database/migration/ ./service/order/database/migration/
COPY service/order_item/database/migration/ ./service/order_item/database/migration/
COPY service/product/database/migration/ ./service/product/database/migration/
COPY service/role/database/migration/ ./service/role/database/migration/
COPY service/transaction/database/migration/ ./service/transaction/database/migration/
COPY service/user/database/migration/ ./service/user/database/migration/

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /app

# Copy source code
COPY pkg/ ./pkg/
COPY shared/ ./shared/
COPY pb/ ./pb/
COPY service/migrate/ ./service/migrate/
COPY docker.env* ./

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/migrate /usr/local/bin/migrate
COPY --from=builder /app/docker.env* ./

COPY --from=builder /app/service ./service/

ENTRYPOINT ["/usr/local/bin/migrate", "up"]
