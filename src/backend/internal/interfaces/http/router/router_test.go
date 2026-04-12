package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Bridge{},
		&model.Drone{},
		&model.Defect{},
		&model.VideoAnalysisTask{},
		&model.DefectObservation{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}

	return db
}

func setupRouterTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Mode: gin.TestMode,
		},
		PythonService: config.PythonServiceConfig{
			Enabled: false,
		},
		Upload: config.UploadConfig{
			ImageDir:  "./test_uploads/images",
			ResultDir: "./test_uploads/results",
			MaxSize:   10,
		},
		Session: config.SessionConfig{
			Secret:     "test-secret",
			MaxAge:     3600,
			CookieName: "test_session",
		},
		CORS: config.CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowCredentials: true,
		},
	}
}

func TestSetupRouter_HealthLegacyAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupRouterTestDB(t)
	r := SetupRouter(db, setupRouterTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSetupRouter_DetectLegacyAliasRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupRouterTestDB(t)
	r := SetupRouter(db, setupRouterTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/detect/image", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
