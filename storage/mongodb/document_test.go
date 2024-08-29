package mongodb

import (
	"context"
	"testing"

	pb "mainService/genproto/docs"

	"github.com/stretchr/testify/assert"
)

// TestConnectMongoDB initializes a connection to MongoDB.
func TestConnectMongoDB(t *testing.T) {
	db, err := ConnectMongoDb() // Ensure this function connects to your test database
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

// TestCreateDocument tests the CreateDocument function.
func TestCreateDocument(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	req := &pb.CreateDocumentReq{
		Title:    "Test Document",
		AuthorId: "test_author",
	}

	res, err := repo.CreateDocument(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, req.Title, res.Title)
	assert.Equal(t, req.AuthorId, res.AuthorId)
}

// TestSearchDocument tests the SearchDocument function.
func TestSearchDocument(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	createReq := &pb.CreateDocumentReq{
		Title:    "Search Test Document",
		AuthorId: "test_author",
	}
	_, _ = repo.CreateDocument(context.Background(), createReq)

	searchReq := &pb.SearchDocumentReq{
		DocsId:   "test_docs_id",
		Title:    createReq.Title,
		AuthorId: createReq.AuthorId,
	}
	res, err := repo.SearchDocument(context.Background(), searchReq)
	assert.NoError(t, err)
	assert.Greater(t, len(res.Documents), 0)
}

// TestGetAllDocuments tests the GetAllDocuments function.
func TestGetAllDocuments(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	req := &pb.GetAllDocumentsReq{
		DocsId:   "test_docs_id",
		AuthorId: "test_author",
	}
	res, err := repo.GetAllDocuments(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Documents)
}

// TestUpdateDocument tests the UpdateDocument function.
func TestUpdateDocument(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	createReq := &pb.CreateDocumentReq{
		Title:    "Update Test Document",
		AuthorId: "test_author",
	}
	createRes, _ := repo.CreateDocument(context.Background(), createReq)

	updateReq := &pb.UpdateDocumentReq{
		Title:    "Updated Document Title",
		Content:  "Updated Content",
		DocsId:   createRes.Title,
		AuthorId: createRes.AuthorId,
	}
	updateRes, err := repo.UpdateDocument(context.Background(), updateReq)
	assert.NoError(t, err)
	assert.Contains(t, updateRes.Message, "Document updated successfully")
}

// TestDeleteDocument tests the DeleteDocument function.
func TestDeleteDocument(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	createReq := &pb.CreateDocumentReq{
		Title:    "Delete Test Document",
		AuthorId: "test_author",
	}
	createRes, _ := repo.CreateDocument(context.Background(), createReq)

	deleteReq := &pb.DeleteDocumentReq{
		Title:    createRes.Title,
		AuthorId: createRes.AuthorId,
	}
	deleteRes, err := repo.DeleteDocument(context.Background(), deleteReq)
	assert.NoError(t, err)
	assert.Equal(t, "Document deleted successfully", deleteRes.Message)
}

// TestShareDocument tests the ShareDocument function.
func TestShareDocument(t *testing.T) {
	db, err := ConnectMongoDb()
	assert.NoError(t, err)
	defer db.Client().Disconnect(context.Background())

	repo := NewDocumentRepository(db)
	createReq := &pb.CreateDocumentReq{
		Title:    "Share Test Document",
		AuthorId: "test_author",
	}
	createRes, _ := repo.CreateDocument(context.Background(), createReq)

	shareReq := &pb.ShareDocumentReq{
		Title:       createRes.Title,
		Id:          createRes.Title,
		UserId:      "collaborator_id",
		Permissions: "read",
	}
	shareRes, err := repo.ShareDocument(context.Background(), shareReq)
	assert.NoError(t, err)
	assert.Equal(t, "Document shared successfully!", shareRes.Message)
}