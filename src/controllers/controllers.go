package controllers

import (
	logger "city_os/src/common"
	"city_os/src/interfaces"
	"city_os/src/models"
	"city_os/src/utils"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

// VideoCatalogueManager, Controller which handled all business logics and talk to DB wrappers
//or any other such Abstractions

type VideoCatalogueManager struct {
	VideoCatalogueDBWrapper interfaces.IDBWrapper
	VideoFilesDBWrapper     interfaces.IFileManagerDBWrapper
}

// GetVideoDocIdBySHAHash, to detect the duplicate video files,
// it is first converting Video file bytes into SHA256 hash
// and performing DB lookup for videos present in the system with the matching hash

func (db *VideoCatalogueManager) GetVideoDocIdBySHAHash(fileDataBytes []byte) (string, string, error) {
	hash, err := utils.ToSHA256(fileDataBytes)
	if err != nil {
		logger.Logger.Error(fmt.Printf("SHA conversion failed!! Error: %s", err.Error()))
		return "", hash, err
	}

	doc, err := db.VideoCatalogueDBWrapper.GetSingleDocByFilter(bson.D{{"hash", hash}})
	if err != nil && !strings.Contains(err.Error(), "no document") {
		logger.Logger.Error(fmt.Printf("Fetching doc by SHA failed!! Error: %s", err.Error()))
		return "", hash, err
	}

	if doc != nil {
		videoCatalogueData := doc.(*models.VideoCatalogueData)
		return videoCatalogueData.FileId, hash, nil
	}
	return "", hash, nil
}

//SaveVideoFile, It is saving video files into the database,
// first creating an entry into  Video Files Meta-Data Storing collection
// then saving the video file bytes into  Video File Bytes Storing Collection in Bytes Chunks (255 KB by default)

func (db *VideoCatalogueManager) SaveVideoFile(
	fileDataBytes []byte,
	filename string,
	fileMimeType string,
	hash string,
) (string, error) {

	videFileCatalogueObj := models.VideoCatalogueData{
		Name:      filename,
		Size:      len(fileDataBytes),
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		FileType:  fileMimeType,
		Hash:      hash,
	}

	docId, err := db.VideoCatalogueDBWrapper.InsertDocument(videFileCatalogueObj)

	if err != nil {
		logger.Logger.Error(fmt.Printf("Insert failed!! Error: %v", err.Error()))
		return "", err
	}

	_, err = db.VideoFilesDBWrapper.UploadFile(docId, fileDataBytes, filename)
	if err != nil {
		logger.Logger.Error(fmt.Printf("Upload file failed!! Error : %v", err.Error()))
		return "", err
	}

	return docId, nil
}

//GetFileByFileId, Fetching Video files data by Video file Id of Document ID of
//Video Files Meta-Data storing collection

func (db *VideoCatalogueManager) GetFileByFileId(
	fileId string,
) (*models.VideoFileData, error) {

	videoCatalogueDataRaw, err := db.VideoCatalogueDBWrapper.GetDocumentById(fileId)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("getDocumentById call failed!! Error:%s", err.Error()))
		return nil, err
	}

	videoCatalogueData := videoCatalogueDataRaw.(*models.VideoCatalogueData)
	videoFileDataBytes, err := db.VideoFilesDBWrapper.DownloadFile(fileId, videoCatalogueData.Name)

	if err != nil || len(videoFileDataBytes) == 0 {
		logger.Logger.Error(fmt.Sprintf("getDocumentById call failed!! Error:%s", err.Error()))
		return nil, err
	}
	fileData := models.VideoFileData{
		videoCatalogueData.Name,
		videoFileDataBytes,
		videoCatalogueData.FileType,
	}

	return &fileData, nil
}

//GetFilesDataById, Fetching Video files meta data from Video Meta-Data storing Collection's Document ID

func (db *VideoCatalogueManager) GetFilesDataById(fileId string) (*models.VideoCatalogueData, error) {
	videoCatalogueDataRaw, err := db.VideoCatalogueDBWrapper.GetDocumentById(fileId)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("getDocumentById call failed!! Error:%s", err.Error()))
		return nil, err
	}

	videoCatalogueData := videoCatalogueDataRaw.(*models.VideoCatalogueData)
	return videoCatalogueData, nil
}

//GetVideoFilesList, Fetching video files list with meta information.

func (db *VideoCatalogueManager) GetVideoFilesList() ([]*models.VideoFilesDataResponse, error) {
	videosListRaw, err := db.VideoCatalogueDBWrapper.GetAllDocuments()
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("getAllDocuments call failed, Error: %s", err.Error()))
		return nil, err
	}
	videosList := make([]*models.VideoFilesDataResponse, 0)
	for _, videoDataRaw := range videosListRaw {
		videoData := videoDataRaw.(*models.VideoCatalogueData)
		videosList = append(videosList, &models.VideoFilesDataResponse{
			videoData.FileId,
			videoData.Name,
			videoData.Size,
			videoData.CreatedAt,
		})
	}
	return videosList, nil
}

//DeleteVideoFile, Deleting video files by Video Meta-Data storing Collection

func (db *VideoCatalogueManager) DeleteVideoFile(fileid string) (bool, error) {
	_, err := db.VideoCatalogueDBWrapper.GetDocumentById(fileid)
	if err != nil {
		return false, err
	}

	_, err = db.VideoCatalogueDBWrapper.DeleteDocumentById(fileid)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Delete doc failed!! Error: %s", err.Error()))
		return false, err
	}

	if err := db.VideoFilesDBWrapper.DeleteFileByFileId(fileid); err != nil {
		logger.Logger.Error(fmt.Sprintf("Doc partially deleted!! Error: %s", err.Error()))
		return false, err
	}
	return true, nil
}
