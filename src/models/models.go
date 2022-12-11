package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VideoCatalogueData struct {
	FileId    string             `bson:"_id,omitempty"` //File Id common for both Meta-Storing and File string Collection
	Name      string             `bson:"name"`          // File name provided by User, We'll be returning this in Get File by Id call
	Size      int                `bson:"size"`          // Video File size in number of Bytes
	CreatedAt primitive.DateTime `bson:"created_at"`    // Video File created at time
	FileType  string             `bson:"type"`          // Video File MIME type
	Hash      string             `bson:"hash"`          // SHA256 hash of Video File Bytes, using it to detect duplicate Video files even with same filename provided
}

type VideoFilesDataResponse struct {
	FileId    string             `json:"fileid,omitempty"`
	Name      string             `json:"name"`
	Size      int                `json:"size"`
	CreatedAt primitive.DateTime `json:"created_at"`
}

type VideoFileData struct {
	Name          string
	FileDataBytes []byte
	FileMimeType  string
}
