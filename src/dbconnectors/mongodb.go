package dbconnectors

import (
	"bytes"
	logger "city_os/src/common"
	"city_os/src/models"
	"city_os/src/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDBClient interface {
	InitConnection(settings interface{})
	GetConnection() interface{}
	GetDBSettings() interface{}
}

type DbClient struct {
	conn interface{}
}

type MongoDBClient struct {
	conn     interface{}
	settings interface{}
}

func (mcli *MongoDBClient) InitConnection(settingsRaw interface{}) {
	settings := settingsRaw.(*MongoDBSettings)
	var client *mongo.Client
	var err error

	opts := options.Client()
	opts.ApplyURI(settings.URI)
	opts.SetMaxPoolSize(settings.PoolSize)
	if client, err = mongo.Connect(context.Background(), opts); err != nil {
		logger.Logger.Fatal("Database connection failed!! Error: %v", err.Error())
	}
	mcli.conn = client
	mcli.settings = settings
}

func (mcli *MongoDBClient) GetConnection() interface{} {
	return mcli.conn
}

func (mcli *MongoDBClient) GetDBSettings() interface{} {
	return mcli.settings
}

type MongoDBSettings struct {
	URI                      string
	PoolSize                 uint64
	VideoCatalogueDB         string
	VideoFilesCollection     string
	VideoCatalogueCollection string
}

type VideoCatalogueDBWrapper struct {
	collection *mongo.Collection
}

func (mdb *VideoCatalogueDBWrapper) InitDatabase(dbClient IDBClient) {
	dbSettings := dbClient.GetDBSettings().(*MongoDBSettings)
	mdb.collection = dbClient.GetConnection().(*mongo.Client).Database(dbSettings.VideoCatalogueDB).Collection(dbSettings.VideoCatalogueCollection)
}

func (mdb *VideoCatalogueDBWrapper) GetDocumentById(id string) (interface{}, error) {
	// convert id string to ObjectId
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var result bson.D
	err = mdb.collection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode(&result)
	if err != nil {
		return nil, err
	}

	docBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, err
	}

	videoCatalogueData := models.VideoCatalogueData{}
	err = bson.Unmarshal(docBytes, &videoCatalogueData)
	if err != nil {
		return nil, err
	}
	return &videoCatalogueData, nil
}

func (mdb *VideoCatalogueDBWrapper) InsertDocument(insertData interface{}) (string, error) {
	insertDocBson, err := utils.ToBson(insertData)

	if err != nil {
		logger.Logger.Fatal(err)
	}

	result, err := mdb.collection.InsertOne(context.Background(), insertDocBson)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (mdb *VideoCatalogueDBWrapper) DeleteDocumentById(id string) (int64, error) {
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		logger.Logger.Println("Invalid id")
	}

	filter := bson.D{{"_id", objectId}}
	result, err := mdb.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		logger.Logger.Fatal(err)
	}
	return result.DeletedCount, nil
}

func (mdb *VideoCatalogueDBWrapper) GetAllDocuments() ([]interface{}, error) {
	cursor, err := mdb.collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}

	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	var videoCatalogueList []interface{}
	for _, result := range results {
		docBytes, err := bson.Marshal(result)
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Bson Marshal failed, result: %v,error: %v", result, err))
		}

		videoCatalogueData := models.VideoCatalogueData{}
		err = bson.Unmarshal(docBytes, &videoCatalogueData)
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Bson UnMarshal failed, result: %v,error: %v", result, err))
		}
		videoCatalogueList = append(videoCatalogueList, &videoCatalogueData)
	}
	return videoCatalogueList, nil
}

func (mdb *VideoCatalogueDBWrapper) GetSingleDocByFilter(filterCondition interface{}) (interface{}, error) {

	var result bson.D
	err := mdb.collection.FindOne(context.Background(), filterCondition).Decode(&result)
	if err != nil {
		return nil, err
	}

	docBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, err
	}

	videoCatalogueData := models.VideoCatalogueData{}
	err = bson.Unmarshal(docBytes, &videoCatalogueData)
	if err != nil {
		return nil, err
	}
	return &videoCatalogueData, nil
}

type VideoFilesDBWrapper struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func (mdb *VideoFilesDBWrapper) InitDatabase(dbClient IDBClient) {
	dbSettings := dbClient.GetDBSettings().(*MongoDBSettings)
	dbConnection := dbClient.GetConnection().(*mongo.Client)

	mdb.database = dbConnection.Database(dbSettings.VideoCatalogueDB)
	mdb.collection = mdb.database.Collection(dbSettings.VideoFilesCollection)
}

func (mdb *VideoFilesDBWrapper) UploadFile(fileID string, fileDataBytes []byte, filename string) (int, error) {
	bucket, err := gridfs.NewBucket(mdb.database)

	if err != nil {
		logger.Logger.Error("GridFS new bucket creation failed!! Error: %v", err)
		return 0, err
	}

	uploadStream, err := bucket.OpenUploadStreamWithID(
		fileID,
		mdb.GetFileNameHash(fileID, filename),
	)

	if err != nil {
		logger.Logger.Error("GridFS opening upload-stream failed!! Error: %v", err)
		return 0, err
	}
	defer uploadStream.Close()

	fileSize, err := uploadStream.Write(fileDataBytes)
	if err != nil {
		logger.Logger.Error("File upload failed!! Error: %v", err)
		return 0, err
	}

	logger.Logger.Info("Write file to DB was successful. File size: %d M\n", fileSize)
	return fileSize, nil
}

func (mdb *VideoFilesDBWrapper) DownloadFile(fileID string, filename string) ([]byte, error) {
	bucket, err := gridfs.NewBucket(
		mdb.database,
	)
	if err != nil {
		logger.Logger.Error(fmt.Printf("bucket creation failed!! Error : %v", err.Error()))
		return nil, err
	}

	var buf bytes.Buffer
	dStream, err := bucket.DownloadToStreamByName(mdb.GetFileNameHash(fileID, filename), &buf)
	if err != nil {
		return nil, err
	}

	fmt.Printf("File size to download: %v\n", dStream)
	return buf.Bytes(), nil
}

func (mdb *VideoFilesDBWrapper) DeleteFileByFileId(fileID string) error {
	bucket, err := gridfs.NewBucket(
		mdb.database,
	)
	if err != nil {
		logger.Logger.Error(fmt.Printf("bucket creation failed!! Error : %v", err.Error()))
		return err
	}
	if err := bucket.Delete(fileID); err != nil {
		return err
	}
	return nil
}

func (mdb *VideoFilesDBWrapper) GetFileNameHash(fileID string, filename string) string {
	return fmt.Sprintf("%s_%s", fileID, filename)
}
