package mongodb

import (
	"context"
	"fmt"
	pb "mainService/genproto/docs"
	"time"

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

	var docs []*pb.GetDocumentRes

	filter := bson.M{
		"authorId": req.AuthorId,
		"title":    req.Title,
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error while finding documents: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc pb.GetDocumentRes
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("error while decoding document: %v", err)
		}
		docs = append(docs, &doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}
	
	return &pb.GetAllDocumentsRes{Documents: docs}, nil
}

func (r *documentVersionRepositoryImpl) RestoreVersion(ctx context.Context, req *pb.RestoreVersionReq) (*pb.RestoreVersionRes, error) {
	coll := r.coll.Collection("docs")

	authorId := req.AuthorId

	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "title", Value: req.Title},
		{Key: "updatedAt", Value: time.Now()},
	}}}

	filter := bson.D{{Key: "authorId", Value: authorId}, {Key: "deletedAt", Value: 0}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &pb.RestoreVersionRes{Message: "Document updated successfully"}, nil
}
