package server

import "time"

// Config defines the common configuration for all gRPC services
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Port           int
	OtelEndpoint   string

	// DBName adalah database MongoDB yang dimiliki service ini.
	// Wajib diisi dengan salah satu konstanta di pkg/database/names.go
	// (IdentityDB, CatalogDB, SalesDB, MerchantDB, EmailDB).
	//
	// Sengaja eksplisit dan tidak diturunkan dari ServiceName: kalau kosong,
	// New() gagal saat start alih-alih diam-diam memakai database default.
	DBName string
}

// Default constants for gRPC server
const (
	DefaultMaxConcurrentConn = 1024
	DefaultWindowSize        = 16 * 1024 * 1024
	DefaultKeepaliveTime     = 20 * time.Second
	DefaultKeepaliveTimeout  = 5 * time.Second
	DefaultMinKeepaliveTime  = 5 * time.Second

	MonitoringInterval     = 30 * time.Second
	CleanupInterval        = 120 * time.Second
	CacheRefCountThreshold = 500

	ShutdownTimeout = 30 * time.Second

	RedisDialTimeout  = 5 * time.Second
	RedisReadTimeout  = 3 * time.Second
	RedisWriteTimeout = 3 * time.Second
	RedisPoolSize     = 10
	RedisMinIdleConns = 3
)
