package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/jeongukjae/pypi-server/internal/db"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

func SetupLegacyRoutes(e *echo.Echo, strg storage.Storage, dbstore db.Store) {
}
