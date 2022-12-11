package interfaces

import "city_os/src/models"

type IDBWrapper interface {
	GetDocumentById(id string) (interface{}, error)
	GetAllDocuments() ([]interface{}, error)
	DeleteDocumentById(id string) (int64, error)
	InsertDocument(insertData interface{}) (string, error)
	GetSingleDocByFilter(filterCondition interface{}) (interface{}, error)
}

type IFileManagerDBWrapper interface {
	UploadFile(fileID string, fileDataBytes []byte, filename string) (int, error)
	DownloadFile(fileName string, filename string) ([]byte, error)
	DeleteFileByFileId(fileID string) error
}

type IVideoCatalogueManager interface {
	GetVideoDocIdBySHAHash(fileDataBytes []byte) (string, string, error)
	SaveVideoFile(
		fileDataBytes []byte,
		filename string,
		fileMimeType string,
		hash string,
	) (string, error)
	GetFileByFileId(
		fileId string,
	) (*models.VideoFileData, error)
	GetFilesDataById(fileId string) (*models.VideoCatalogueData, error)
	GetVideoFilesList() ([]*models.VideoFilesDataResponse, error)
	DeleteVideoFile(fileid string) (bool, error)
}
