package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jeongukjae/pypi-server/internal/storage"
)

func SetupSimpleRoutes(e *echo.Echo, strg storage.Storage) {
	e.GET("/simple/", ListPackages(strg))
	e.GET("/simple/:package/", ListPackageFiles(strg))
	e.GET("/simple/:package/:file", DownloadFile(strg))
}

func ListPackages(strg storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		packages, err := strg.ListPackages(c.Request().Context())
		if err != nil {
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

func ListPackageFiles(strg storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		packageName := c.Param("package")
		files, err := strg.ListPackageFiles(c.Request().Context(), packageName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &HTTPError{Message: "Failed to list package files", Errors: []string{err.Error()}})
		}

		html := "<!DOCTYPE html><html><body>"
		for _, file := range files {
			html += `<a href="/simple/` + packageName + `/` + file + `">` + file + `</a><br/>`
		}
		html += "</body></html>"

		return c.HTML(http.StatusOK, html)
	}
}

func DownloadFile(strg storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		packageName := c.Param("package")
		fileName := c.Param("file")

		rc, err := strg.ReadFile(c.Request().Context(), packageName, fileName)
		if err != nil {
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
