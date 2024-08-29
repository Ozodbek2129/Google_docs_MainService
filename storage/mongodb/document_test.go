package mongodb

import (
	"context"
	"errors"
	"testing"

	pb "mainService/genproto/docs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MockMongoCollection is a mock type for the Mongo Collection
type MockMongoCollection struct {
	mock.Mock
}

// Mock functions for collection methods
func (m *MockMongoCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

// MockMongoDatabase is a mock type for the Mongo Database
type MockMongoDatabase struct {
	mock.Mock
	*mongo.Database
}

func (m *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *MockMongoCollection {
	args := m.Called(name)
	return args.Get(0).(*MockMongoCollection)
}

func TestCreateDocument(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)

	// Setting expectations
	mockDB.On("Collection", "docs").Return(mockCollection)
	mockCollection.On("FindOne", mock.Anything, mock.Anything).
		Return(&mongo.SingleResult{}) // Mock FindOne return value
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).
		Return(&mongo.InsertOneResult{InsertedID: "mockID"}, nil) // Mock InsertOne return value
		
	repo := NewDocumentRepository(mockDB.Database)
	req := &pb.CreateDocumentReq{
		Title:    "Test Document",
		AuthorId: "author123",
	}

	// Call CreateDocument
	res, err := repo.CreateDocument(context.Background(), req)

	// Assert no error occurred
	assert.NoError(t, err)
	assert.Equal(t, "Test Document", res.Title)
	assert.Equal(t, "author123", res.AuthorId)

	// Assert expectations
	mockDB.AssertExpectations(t)
	mockCollection.AssertExpectations(t)

}

func TestCreateDocument_DuplicateTitle(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)

	// Setting expectations
	mockDB.On("Collection", "docs").Return(mockCollection)
	mockCollection.On("FindOne", mock.Anything, mock.Anything).
		Return(&mongo.SingleResult{}) // Mock FindOne return value

	mockCollection.On("InsertOne", mock.Anything, mock.Anything).
		Return(&mongo.InsertOneResult{}, mongo.IsDuplicateKeyError(errors.New("duplicate key error")))

	repo := NewDocumentRepository(mockDB.Database)

	req := &pb.CreateDocumentReq{
		Title:    "Duplicate Title",
		AuthorId: "author123",
	}

	// Call CreateDocument
	_, err := repo.CreateDocument(context.Background(), req)

	// Assert error occurred and check the error message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title 'Duplicate Title' is already taken")

	// Assert expectations
	mockDB.AssertExpectations(t)
	mockCollection.AssertExpectations(t)
}