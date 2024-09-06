package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	pb "mainService/genproto/doccs"
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
	var existingDoc bson.M

	err := coll.FindOne(ctx, bson.M{"authorId": req.AuthorId}).Decode(&existingDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	var docsId string
	if err == mongo.ErrNoDocuments {
		docsId = uuid.NewString()
	} else {
		docsId = existingDoc["docsId"].(string)
	}

	_, err = coll.InsertOne(ctx, bson.M{
		"_id":            uuid.New().String(),
		"title":          req.Title,
		"content":        "",
		"docsId":         docsId,
		"authorId":       req.AuthorId,
		"collaboratorId": "",
		"version":        0,
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

	filter := bson.M{
		"docsId": req.DocsId,
		"title":  req.Title,
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*pb.GetDocumentRes

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		if doc["authorId"] == req.AuthorId {
			result := &pb.GetDocumentRes{
				Title:          doc["title"].(string),
				Content:        doc["content"].(string),
				DocsId:         doc["docsId"].(string),
				AuthorId:       doc["authorId"].(string),
				Version: 		doc["version"].(int32),	
			}
			results = append(results, result)
		} else {
			collaboratorIds, ok := doc["collaboratorId"].([]string)
			if !ok || !contains(collaboratorIds, req.AuthorId) {
				continue
			}

			result := &pb.GetDocumentRes{
				Title:          doc["title"].(string),
				Content:        doc["content"].(string),
				DocsId:         doc["docsId"].(string),
				AuthorId:       doc["authorId"].(string),
				Version:        doc["version"].(int32),
			}
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("authorId '%s' hujjatlar topilmadi", req.AuthorId)
	}

	return &pb.SearchDocumentRes{Documents: results}, nil
}

func contains(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func (r *documentRepositoryImpl) GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error) {
	coll := r.coll.Collection("docs")

	if req.DocsId == "" {
		return nil, fmt.Errorf("docs id '%s' is not set", req.DocsId)
	}

	filter := bson.M{
		"docsId": req.DocsId,
	}

	findOptions := options.Find()

	if req.Limit != 0 {
		findOptions.SetLimit(int64(req.Limit))
	} else {
		findOptions.SetLimit(10)
	}

	if req.Page != 0 {
		findOptions.SetSkip(int64((req.Page - 1) * int32(req.Limit)))
	} else {
		findOptions.SetSkip(0)
	}

	cursor, err := coll.Find(ctx, filter, findOptions)
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
				Title:          doc["title"].(string),
				Content:        doc["content"].(string),
				DocsId:         doc["docsId"].(string),
				AuthorId:       doc["authorId"].(string),
				Version:        doc["version"].(int32),
			}
			docs = append(docs, result)
			authorFound = true
		} else {
			collaboratorIds, ok := doc["collaboratorId"].([]string)
			if ok && contains(collaboratorIds, req.AuthorId) {
				result := &pb.GetDocumentRes{
					Title:          doc["title"].(string),
					Content:        doc["content"].(string),
					DocsId:         doc["docsId"].(string),
					AuthorId:       doc["authorId"].(string),
					Version:        doc["version"].(int32),
				}
				docs = append(docs, result)
				authorFound = true
			}
		}
	}

	if !authorFound {
		return nil, fmt.Errorf("no documents found for docsId '%s' with the given authorId or in collaboratorId", req.DocsId)
	}

	return &pb.GetAllDocumentsRes{Documents: docs}, nil
}


func (r *documentRepositoryImpl) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error) {
	coll := r.coll.Collection("docs")

	if req.AuthorId == "" {
		return nil, fmt.Errorf("authorId '%s' is not set", req.AuthorId)
	}

	if req.DocsId == ""{
		return nil,fmt.Errorf("docs id '%s' is not set",req.DocsId)
	}

	filter := bson.D{{Key: "authorId", Value: req.AuthorId}, {Key: "docsId", Value: req.DocsId}, {Key: "deletedAt", Value: 0}}

	var existingDoc bson.M
	err := coll.FindOne(ctx, filter).Decode(&existingDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document with authorId '%s' and docsId '%s' not found", req.AuthorId, req.DocsId)
		}
		return nil, err
	}

	newVersion := existingDoc["version"].(int32) + 1

	_, err = coll.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "deletedAt", Value: time.Now().Unix()},
		}},
	})
	if err != nil {
		return nil, err
	}

	_, err = coll.InsertOne(ctx, bson.M{
		"_id":            uuid.New().String(),
		"title":          req.Title,
		"content":        req.Content,
		"docsId":         req.DocsId,
		"authorId":       req.AuthorId,
		"collaboratorId": existingDoc["collaboratorId"].(string),
		"version":        newVersion,
		"createdAt":      existingDoc["createdAt"],
		"updatedAt":      time.Now(),
		"deletedAt":      0,
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateDocumentRes{Message: "Document updated successfully with version " + fmt.Sprintf("%d", newVersion)}, nil
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
		"_id":   req.Id,
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
