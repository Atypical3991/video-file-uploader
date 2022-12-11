package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

func ToBson(v interface{}) (*bson.D, error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}
	doc := bson.D{}
	err = bson.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func ToSHA256(data []byte) (hash string, error error) {
	defer func() {
		if err := recover(); err != nil {
			error = errors.New("panic occurred on hash creation")
		}
	}()

	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}
