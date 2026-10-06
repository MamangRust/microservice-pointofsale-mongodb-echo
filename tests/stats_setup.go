package tests

import (
	"context"
	"fmt"
	"time"

	pkgclickhouse "github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// statsClickHouseDB is the database the stats suites use inside the ClickHouse
// testcontainer. It is created by EnsureDatabase so the reader/writer connect
// with it as their default database.
const statsClickHouseDB = "pos_stats_test"

// The official ClickHouse image's `default` user is only reachable from
// localhost, so a client connecting through the mapped port is rejected as an
// authentication failure. The container is started with a dedicated user that
// is allowed from any network, and viper is pointed at the same credentials.
const (
	statsClickHouseUser     = "stats"
	statsClickHousePassword = "stats"
)

// statsClickHouseImage is the ClickHouse server image used by the stats
// integration suites. Kept in sync with deployments/local/docker-compose.infra.yml.
const statsClickHouseImage = "clickhouse/clickhouse-server:24.8-alpine"

// SetupStatsEnv starts a ClickHouse testcontainer and points viper at it so the
// stats-writer (backfill) and stats-reader suites exercise a real OLAP store.
// Call it after BaseTestSuite.SetupSuite; the container is terminated in
// TearDownSuite.
func (s *BaseTestSuite) SetupStatsEnv() {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        statsClickHouseImage,
			ExposedPorts: []string{"9000/tcp", "8123/tcp"},
			Env: map[string]string{
				"CLICKHOUSE_USER":                      statsClickHouseUser,
				"CLICKHOUSE_PASSWORD":                  statsClickHousePassword,
				"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
			},
			WaitingFor: wait.ForHTTP("/ping").
				WithPort("8123/tcp").
				WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	s.Require().NoError(err)
	s.statsContainer = container

	host, err := container.Host(ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(ctx, "9000")
	s.Require().NoError(err)

	viper.Set("CLICKHOUSE_ADDR", fmt.Sprintf("%s:%s", host, port.Port()))
	viper.Set("CLICKHOUSE_DATABASE", statsClickHouseDB)
	viper.Set("CLICKHOUSE_USERNAME", statsClickHouseUser)
	viper.Set("CLICKHOUSE_PASSWORD", statsClickHousePassword)
	// Backfill no longer opens Mongo itself, but keep the URI wired for any
	// suite component that still resolves it through viper.
	viper.Set("MONGO_URI", s.MongoURI())

	s.Require().NoError(pkgclickhouse.EnsureDatabase(s.Log))
}
