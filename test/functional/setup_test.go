// Package functional contains end-to-end tests for the Notification & Messaging Service.
// These tests are allowed to connect to a real database and external services.
// They are executed in pipeline step 5 (after Build Image) against a local or staging environment.
//
// Run with:
//   go test ./test/functional/... -v -tags=functional -timeout 120s
package functional

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testEnv holds shared state for the functional test suite
type testEnv struct {
	db  *gorm.DB
	ctx context.Context
}

// setupTestEnv initialises the test environment.
// It reads connection strings from environment variables so that both local
// Docker-compose and CI/CD environments are supported without code changes.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dsn := os.Getenv("FUNCTIONAL_TEST_DB_DSN")
	if dsn == "" {
		// Default DSN for local docker-compose setup
		dsn = "host=localhost user=notif_user password=notif_pass dbname=notification_db port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "gagal koneksi ke database; pastikan service postgres sudah berjalan")

	env := &testEnv{
		db:  db,
		ctx: context.Background(),
	}

	runMigrations(t, db)
	return env
}

// teardownTestEnv cleans up test data after each test
func teardownTestEnv(t *testing.T, env *testEnv) {
	t.Helper()
	cleanTestData(t, env.db)
}

// runMigrations runs schema migrations required for testing
func runMigrations(t *testing.T, db *gorm.DB) {
	t.Helper()
	// TODO: panggil AutoMigrate atau jalankan migration scripts
	// db.AutoMigrate(&model.Notification{}, &model.MessageTemplate{}, &model.UserContact{})
}

// cleanTestData removes rows inserted during tests (prefix "test-")
func cleanTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Hapus data test agar idempotent
	db.Exec("DELETE FROM notifications WHERE user_id LIKE 'ft-user-%'")
	db.Exec("DELETE FROM message_templates WHERE name LIKE 'ft_%'")
}
