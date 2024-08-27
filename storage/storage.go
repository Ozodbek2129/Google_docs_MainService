package storage

import (
	pb "mainService/genproto/docs"
)

type IStorage interface {
	Docs() IDocsStorage
	Close()
}

type IDocsStorage interface {
	CreateDocument(req *pb.CreateDocumentReq) (*pb.CreateDocumentRes, error)
	GetDocument(req *pb.GetDocumentReq) (*pb.GetDocumentRes, error)
	GetAllDocuments(req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error)
	UpdateDocument(req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error)
	DeleteDocument(req *pb.DeleteDocumentReq) (*pb.DeleteDocumentRes, error)
	ShareDocument(req *pb.ShareDocumentReq) (*pb.ShareDocumentRes, error)
	SearchDocument(req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error)
	GetAllVersions(req *pb.GetAllVersionsReq) (*pb.GetAllVersionsRes, error)
	RestoreVersion(req *pb.RestoreVersionReq) (*pb.RestoreVersionRes, error)
}