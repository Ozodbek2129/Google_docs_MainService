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
		"_id":   id,
        "title": req.Title,
        "content": req.Content,
		"tags":     req.Tags,
        "category": req.Category,
        "authorId":   req.AuthorId,
		"collobratorId": req.CollabratorId,
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
		"deletedAt": 0,
	})

	if err!= nil {
        return nil, err
    }

	return &pb.CreateDocumentRes{Message: "Document created successfully"}, nil
}