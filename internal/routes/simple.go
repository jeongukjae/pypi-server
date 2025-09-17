package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/packageindex"
)

func SetupSimpleRoutes(e *echo.Echo, index packageindex.Index) {
	e.GET("/simple/", ListPackages(index))
	e.GET("/simple/:package/", ListPackageFiles(index))
	e.GET("/simple/:package/:file", DownloadFile(index))
}

// TODO: To support JSON API, add negotiation for Accept header

func ListPackages(index packageindex.Index) echo.HandlerFunc {
	return func(c echo.Context) error {
		packages, err := index.ListPackages(c.Request().Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to list packages from database")
			return c.JSON(http.StatusInternalServerError, &HTTPError{Message: "Failed to list packages", Errors: []string{err.Error()}})
		}

		html := "<!DOCTYPE html><html><body>"
		for _, pkg := range packages {
			html += `<a href="/simple/` + pkg + `/">` + pkg + `</a>`
		}
		html += "</body></html>"

		return c.HTML(http.StatusOK, html)
	}
}

func ListPackageFiles(index packageindex.Index) echo.HandlerFunc {
	return func(c echo.Context) error {
		packageName := c.Param("package")

		html := "<!DOCTYPE html><html><body>"

		files, err := index.ListPackageFiles(c.Request().Context(), packageName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list package files")
			return c.JSON(http.StatusInternalServerError, &HTTPError{Message: "Failed to list package files", Errors: []string{err.Error()}})
		}
		for _, file := range files {
			html += `<a href="/simple/` + packageName + `/` + file + `">` + file + `</a><br/>`
		}
		html += "</body></html>"

		return c.HTML(http.StatusOK, html)
	}
}

func DownloadFile(index packageindex.Index) echo.HandlerFunc {
	return func(c echo.Context) error {
		packageName := c.Param("package")
		fileName := c.Param("file")

		rc, err := index.DownloadFile(c.Request().Context(), packageName, fileName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read file")
			return c.JSON(http.StatusInternalServerError, &HTTPError{Message: "Failed to read file", Errors: []string{err.Error()}})
		}
		defer rc.Close()

		return c.Stream(
			http.StatusOK,
			"application/octet-stream",
			rc,
		)
	}
}
