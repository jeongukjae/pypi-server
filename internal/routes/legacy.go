package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/jeongukjae/pypi-server/internal/packageindex"
)

func SetupLegacyRoutes(e *echo.Echo, index packageindex.Index) {
	e.POST("/legacy/", UploadFile(index))
}

// Reference:
// - https://docs.pypi.org/api/upload/
// - https://peps.python.org/pep-0694/

// TODO: need to support negotiation for Content-Type header
// TODO: need to match error response as stated in PEP 694

type UploadFilePayload struct {
	Action          string `form:":action"`
	ProtocolVersion string `form:"protocol_version"`
	Version         string `form:"version"`
	Name            string `form:"name"`
	FileType        string `form:"filetype"`
	MetadataVersion string `form:"metadata_version"`

	Summary                *string  `form:"summary"`
	Description            *string  `form:"description"`
	DescriptionContentType *string  `form:"description_content_type"`
	PyVersion              *string  `form:"pyversion"`
	RequiresDist           []string `form:"requires_dist"`

	Md5Digest        *string `form:"md5_digest"`
	Sha256Digest     *string `form:"sha256_digest"`
	Blake2_256Digest *string `form:"blake2_256_digest"`
	// Attestation not supported yet
}

func UploadFile(index packageindex.Index) echo.HandlerFunc {
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

		formFile, err := c.FormFile("content")
		if err != nil {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{"File is required"}})
		}

		file, err := formFile.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, &HTTPError{Message: "Invalid request", Errors: []string{"Failed to open uploaded file"}})
		}
		defer file.Close()

		log.Ctx(c.Request().Context()).Debug().Str("package", payload.Name).Str("version", payload.Version).Str("file", formFile.Filename).Msg("Uploading file")

		err = index.UploadFile(c.Request().Context(), packageindex.UploadFileRequest{
			PackageName:            payload.Name,
			Version:                payload.Version,
			FileName:               formFile.Filename,
			FileType:               payload.FileType,
			MetadataVersion:        payload.MetadataVersion,
			Summary:                payload.Summary,
			Description:            payload.Description,
			DescriptionContentType: payload.DescriptionContentType,
			Pyversion:              payload.PyVersion,
			RequiresPython:         payload.PyVersion,
			RequiresDist:           payload.RequiresDist,
			Md5Digest:              payload.Md5Digest,
			Sha256Digest:           payload.Sha256Digest,
			Blake2256Digest:        payload.Blake2_256Digest,
		}, file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &HTTPError{Message: "Failed to upload file", Errors: []string{err.Error()}})
		}

		log.Ctx(c.Request().Context()).Info().Str("package", payload.Name).Str("version", payload.Version).Str("file", formFile.Filename).Msg("File uploaded")
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	}
}
