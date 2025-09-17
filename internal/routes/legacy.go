package routes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jeongukjae/pypi-server/internal/db"
	"github.com/jeongukjae/pypi-server/internal/storage"
)

func SetupLegacyRoutes(e *echo.Echo, strg storage.Storage, dbstore db.Store) {
}

// Reference: https://docs.pypi.org/api/upload/

type UploadFilePayload struct {
	Action           string  `form:":action"`
	ProtocolVersion  string  `form:"protocol_version"`
	Name             string  `form:"name"`
	Summary          *string `form:"summary"`
	Md5Digest        *string `form:"md5_digest"`
	Sha256Digest     *string `form:"sha256_digest"`
	Blake2_256Digest *string `form:"blake2_256_digest"`
	FileType         string  `form:"filetype"`
	PyVersion        string  `form:"pyversion"`
	MetadataVersion  string  `form:"metadata_version"`

	// Attestation not supported yet
}

func UploadFile(strg storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload UploadFilePayload
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{err.Error()}})
		}

		if payload.Action != "file_upload" {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{"Invalid action, only file_upload is supported"}})
		}

		if payload.ProtocolVersion != "1" {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{"Invalid protocol_version, only version 1 is supported"}})
		}

		fmt.Printf("payload: %+v\n", payload)

		// print other form fields
		formParams, err := c.FormParams()
		if err != nil {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{err.Error()}})
		}

		for key, values := range formParams {
			if key != "file" {
				fmt.Printf("Form field: %s = %v (len: %d)\n", key, values, len(values))
			}
		}

		return c.JSON(http.StatusNotImplemented, &HTTPError{Message: "Not implemented", Errors: []string{"File upload not implemented yet"}})
	}
}
