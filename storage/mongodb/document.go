package mongodb

import (
	"context"
	pb "mainService/genproto/docs"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, req *pb.CreateDocumentReq) (*pb.CreateDocumentRes, error)
	SearchDocument(ctx context.Context, req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error)
	GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error)
	UpdateDocument(ctx context.Context, req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error)
	DeleteDocument(ctx context.Context, req *pb.DeleteDocumentReq) (*pb.DeleteDocumentRes, error)
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
		"_id":           id,
		"title":         req.Title,
		"content":       req.Content,
		"tags":          req.Tags,
		"category":      req.Category,
		"authorId":      req.AuthorId,
		"collobratorId": "",
		"createdAt":     time.Now(),
		"updatedAt":     time.Now(),
		"deletedAt":     0,
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreateDocumentRes{Message: "Document created successfully"}, nil
}

func (r *documentRepositoryImpl) SearchDocument(ctx context.Context, req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error) {
	coll := r.coll.Collection("docs")

	var doc pb.SearchDocumentRes
	err := coll.FindOne(ctx, bson.M{"_id": req.Title}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	err = coll.FindOne(ctx, bson.M{"_id": req.AuthorId}).Decode(&doc)

	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepositoryImpl) GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error) {
	coll := r.coll.Collection("docs")

	filtr := bson.M{}
	if req.Email != "" {
		filtr["authorId"] = req.Email
	}

	if req.Limit != 0 {
		filtr["limit"] = req.Limit
	}
	if req.Page != 0 {
		filtr["page"] = req.Page
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

	authorId := req.AuthorId

	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "title", Value: req.Title},
		{Key: "content", Value: req.Content},
		{Key: "tags", Value: req.Tags},
		{Key: "category", Value: req.Category},
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

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now().Unix()}}}}

	filter := bson.D{{Key: "authorId", Value: req.AuthorId}, {Key: "deletedAt", Value: 0}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return &pb.DeleteDocumentRes{
			Message: "Document did not deleted",
		}, err
	}

	return &pb.DeleteDocumentRes{Message: "Document deleted successfully"}, nil
}
