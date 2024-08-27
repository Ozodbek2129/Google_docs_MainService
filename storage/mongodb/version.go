package mongodb

import (
	"context"
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

	var doc []*pb.GetDocumentRes
	err := coll.FindOne(ctx, bson.M{"_id": req.Title}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	err = coll.FindOne(ctx, bson.M{"_id": req.AuthorId}).Decode(&doc)

	if err != nil {
		return nil, err
	}
	return &pb.GetAllDocumentsRes{Documents: doc}, nil
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
