package mongodb

import (
	"context"
	"fmt"
	pb "mainService/genproto/docs"
	"time"

	"github.com/syndtr/goleveldb/leveldb/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentVersionRepository interface {
	GetAllVersions(ctx context.Context, req *pb.GetAllVersionsReq) (*pb.GetAllDocumentsRes, error)
	RestoreVersion(ctx context.Context, req *pb.RestoreVersionReq) (*pb.RestoreVersionRes, error)
}

type documentVersionRepositoryImpl struct {
	coll *mongo.Database
}

func NewDocumentVersionRepository(db *mongo.Database) DocumentVersionRepository {
	return &documentVersionRepositoryImpl{coll: db}
}

func (r *documentVersionRepositoryImpl) GetAllVersions(ctx context.Context, req *pb.GetAllVersionsReq) (*pb.GetAllDocumentsRes, error) {
	coll := r.coll.Collection("docs")

	if req.AuthorId == "" {
		return nil, errors.New("Author is required")
	}
	if req.Title == "" {
		return nil, errors.New("Title is required")
	}

	filter := bson.M{
		"authorId": req.AuthorId,
		"title":    req.Title,
	}

	cursor, err := coll.Find(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*pb.GetDocumentRes
	var authorFound bool

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		authorId, ok := doc["authorId"].(string)
		if !ok {
			continue
		}

		if authorId == req.AuthorId {
			result := &pb.GetDocumentRes{
				Title:    doc["title"].(string),
				Content:  doc["content"].(string),
				DocsId:   doc["docsId"].(string),
				AuthorId: doc["authorId"].(string),
				Version:  doc["version"].(int32),
			}
			docs = append(docs, result)
			authorFound = true
		} else {
			collaboratorIds, ok := doc["collaboratorId"].([]string)
			if ok && contains(collaboratorIds, req.AuthorId) {
				result := &pb.GetDocumentRes{
					Title:    doc["title"].(string),
					Content:  doc["content"].(string),
					DocsId:   doc["docsId"].(string),
					AuthorId: doc["authorId"].(string),
					Version:  doc["version"].(int32),
				}
				docs = append(docs, result)
				authorFound = true
			}
		}
	}

	if !authorFound {
		return nil, fmt.Errorf("no documents found '%s' with the given authorId or in collaboratorId", req.AuthorId)
	}

	return &pb.GetAllDocumentsRes{Documents: docs}, nil
}

func (r *documentVersionRepositoryImpl) RestoreVersion(ctx context.Context, req *pb.RestoreVersionReq) (*pb.RestoreVersionRes, error) {
	coll := r.coll.Collection("docs")

	authorId := req.AuthorId

	if authorId == "" {
		return nil, errors.New("Author is required")
	}
	if req.Title == "" {
		return nil, errors.New("Title is required")
	}
	if req.Version == 0 {
        return nil, errors.New("Version is required")
    }

	filter := bson.D{
		{Key: "authorId", Value: authorId},
		{Key: "title", Value: req.Title},
		{Key: "version", Value: req.Version},
		{Key: "docsId", Value: req.Id},
	}

	var doc pb.GetDocumentRes
	err := coll.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("document with docsId '%s' and title '%s' not found", req.Id, req.Title)
	}

	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "title", Value: req.Title},
		{Key: "content", Value: doc.Content},
		{Key: "updatedAt", Value: time.Now()},
	}}}

	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &pb.RestoreVersionRes{Message: "Document updated successfully"}, nil
}
