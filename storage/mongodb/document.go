package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	pb "mainService/genproto/docs"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, req *pb.CreateDocumentReq) (*pb.CreateDocumentRes, error)
	SearchDocument(ctx context.Context, req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error)
	GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error)
	UpdateDocument(ctx context.Context, req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error)
	DeleteDocument(ctx context.Context, req *pb.DeleteDocumentReq) (*pb.DeleteDocumentRes, error)
	ShareDocument(ctx context.Context, req *pb.ShareDocumentReq) (*pb.ShareDocumentRes, error)
}

type documentRepositoryImpl struct {
	coll *mongo.Database
}

func NewDocumentRepository(db *mongo.Database) DocumentRepository {
	return &documentRepositoryImpl{coll: db}
}

func (r *documentRepositoryImpl) CreateDocument(ctx context.Context, req *pb.CreateDocumentReq) (*pb.CreateDocumentRes, error) {
	coll := r.coll.Collection("docs")
	id := uuid.New().String()

	_, err := coll.InsertOne(ctx, bson.M{
		"_id":            id,
		"title":          req.Title,
		"content":        req.Content,
		"authorId":       req.AuthorId,
		"collaboratorId": "",
		"version":        1,
		"createdAt":      time.Now(),
		"updatedAt":      time.Now(),
		"deletedAt":      0,
	})

	if mongo.IsDuplicateKeyError(err) {
		return nil, fmt.Errorf("title '%s' is already taken", req.Title)
	} else if err != nil {
		return nil, err
	}

	return &pb.CreateDocumentRes{Title: req.Title, AuthorId: req.AuthorId}, nil
}

func (r *documentRepositoryImpl) SearchDocument(ctx context.Context, req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error) {
	coll := r.coll.Collection("docs")

	var doc pb.SearchDocumentRes
	filter := bson.M{}

	if req.Title != "" {
		filter["title"] = req.Title
	} else {
		return nil, fmt.Errorf("title '%s' is not set", req.Title)
	}

	if req.AuthorId != "" {
		filter["authorId"] = req.AuthorId
	} else {
		return nil, fmt.Errorf("authorId '%s' is not set", req.AuthorId)
	}

	err := coll.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *documentRepositoryImpl) GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error) {
	coll := r.coll.Collection("docs")

	if req.AuthorId == "" {
		return nil, fmt.Errorf("authorId '%s' is not set", req.AuthorId)
	}

	filtr := bson.M{
		"authorId": req.AuthorId,
	}

	if req.Limit != 0 {
		filtr["limit"] = req.Limit
	} else {
		filtr["limit"] = 10
	}

	if req.Page != 0 {
		filtr["page"] = req.Page
	} else {
		filtr["page"] = 1
	}

	cursor, err := coll.Find(ctx, filtr)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []*pb.GetDocumentRes
	for cursor.Next(ctx) {
		var doc pb.GetDocumentRes
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		docs = append(docs, &doc)
	}
	return &pb.GetAllDocumentsRes{Documents: docs}, nil
}

func (r *documentRepositoryImpl) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error) {
	coll := r.coll.Collection("docs")

	if req.AuthorId != "" {
		return nil, fmt.Errorf("authorId '%s' is not set", req.AuthorId)
	}

	authorId := req.AuthorId

	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "title", Value: req.Title},
		{Key: "content", Value: req.Content},
		{Key: "updatedAt", Value: time.Now()},
	}}}

	filter := bson.D{{Key: "authorId", Value: authorId}, {Key: "deletedAt", Value: 0}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateDocumentRes{Message: "Document updated successfully"}, nil
}

func (r *documentRepositoryImpl) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentReq) (*pb.DeleteDocumentRes, error) {
	coll := r.coll.Collection("docs")

	filter := bson.D{
		{Key: "authorId", Value: req.AuthorId},
		{Key: "title", Value: req.Title},
		{Key: "deletedAt", Value: 0},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "deletedAt", Value: time.Now().Unix()},
		}},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil || result.MatchedCount == 0 {
		return &pb.DeleteDocumentRes{
			Message: "Document deletion failed",
		}, fmt.Errorf("failed to delete document: %w", err)
	}

	return &pb.DeleteDocumentRes{
		Message: "Document deleted successfully",
	}, nil
}

func (r *documentRepositoryImpl) ShareDocument(ctx context.Context, req *pb.ShareDocumentReq) (*pb.ShareDocumentRes, error) {
	coll := r.coll.Collection("docs")
	filter := bson.M{
		"title": req.Title,
		"_id": req.Id,
	}

	var existingDoc bson.M
	err := coll.FindOne(ctx, filter).Decode(&existingDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error while finding document: %v", err)
	}

	newCollaboration := bson.M{
		req.UserId: req.Permissions,
	}

	if existingCollaboration, ok := existingDoc["collaboratorId"].(bson.M); ok {
		for key, value := range newCollaboration {
			existingCollaboration[key] = value
		}
		newCollaboration = existingCollaboration
	}

	collaborationBytes, err := json.Marshal(newCollaboration)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling collaboration data: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"collaboratorId": string(collaborationBytes),
		},
	}

	_, err = coll.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("error while updating document: %v", err)
	}

	res := &pb.ShareDocumentRes{
		Message: "Document shared successfully!",
	}

	return res, nil
}
