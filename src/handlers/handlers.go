package handlers

import (
	"bytes"
	"city_os/cmd/app/configs"
	logger "city_os/src/common"
	"city_os/src/interfaces"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type Handler struct {
	VideoCatalogueManager interfaces.IVideoCatalogueManager
	Config                *configs.AppConfig
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) GetFileByIdHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error occurred"})
		}
	}()
	fileid, found := c.Params.Get("fileid")
	if !found {
		logger.Logger.Info(fmt.Sprintf("Bad Request!!"))
		c.JSON(http.StatusBadRequest, gin.H{"message": "fileID is a mandatory path param!!"})
		return
	}

	fileData, err := h.VideoCatalogueManager.GetFileByFileId(fileid)
	if err != nil {
		logger.Logger.Info(fmt.Sprintf("File not found!! fileID:%s", fileid))
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found!!", "error": err.Error()})
		return
	}

	responseWriter := c.Writer
	responseWriter.Header().Set("Content-Type", fileData.FileMimeType)
	responseWriter.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileData.Name))
	if _, err = io.Copy(responseWriter, bytes.NewReader(fileData.FileDataBytes)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "IO write into Response failed", "error": err.Error()})
	}
}

func (h *Handler) LocateFileByIdHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error occurred"})
		}
	}()

	fileid, found := c.Params.Get("fileid")
	if !found {
		logger.Logger.Info(fmt.Sprintf("Bad Request!!"))
		c.JSON(http.StatusBadRequest, gin.H{"message": "fileID is a mandatory path param!!"})
		return
	}

	fileData, err := h.VideoCatalogueManager.GetFilesDataById(fileid)
	if err != nil {
		logger.Logger.Info(fmt.Sprintf("File not found!! fileID:%s", fileid))
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found!!", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fileData": fileData})
}

func (h *Handler) DeleteFileByIdHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error occurred"})
		}
	}()
	fileId, found := c.Params.Get("fileid")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"message": "fileID is a mandatory path param"})
		return
	}

	_, err := h.VideoCatalogueManager.DeleteVideoFile(fileId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found", "error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *Handler) PostSingleFileHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error occurred"})
		}
	}()
	//Supported media types
	supportedMediaTypes := []string{`video/mp4`, `video/mpeg`}
	file, header, err := c.Request.FormFile("data")
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Parsing form-data failed!! Error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Parsing form-data failed", "error": err.Error()})
		return
	}
	contentType := header.Header.Get("Content-Type")
	isSupportedMediaType := false
	for _, mType := range supportedMediaTypes {
		if mType == contentType {
			isSupportedMediaType = true
			break
		}
	}

	if !isSupportedMediaType {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "media type not supported"})
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		logger.Logger.Error(fmt.Printf("Video file byte conversion failed!! Error: %v", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Video file parsing failed.", "error": err.Error()})
		return
	}

	docId, hash, err := h.VideoCatalogueManager.GetVideoDocIdBySHAHash(buf.Bytes())
	if err != nil {
		logger.Logger.Error(fmt.Printf("Fetchnig doc by hash failed!! Error: %v", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Hash conversion failed!!", "error": err.Error()})
		return
	}

	if docId != "" {
		logger.Logger.Info(fmt.Printf("Duplicate doc found!! docId : %s", docId))
		c.JSON(http.StatusConflict, gin.H{"message": fmt.Sprintf("File exists!! docId : %s", docId)})
		return
	}

	fileDocId, err := h.VideoCatalogueManager.SaveVideoFile(buf.Bytes(), header.Filename, contentType, hash)
	if err != nil {
		logger.Logger.Error(fmt.Printf("Video file byte conversion failed!! Error: %v", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Video file parsing failed.", "error": err.Error()})
		return
	}
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	c.Redirect(http.StatusCreated, fmt.Sprintf("http://%s:%s/v1/files/locate/%s", host, port, fileDocId))
}

func (h *Handler) GetFilesListHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error occurred"})
		}
	}()
	videosList, err := h.VideoCatalogueManager.GetVideoFilesList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Fetching videos list failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videosList)
}
