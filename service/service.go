package service

import (
	"context"
	"log/slog"
	pb "mainService/genproto/doccs"
	"mainService/storage/mongodb"
)

type Service struct {
	pb.UnimplementedDocsServiceServer
	logger *slog.Logger
	repo   mongodb.DocumentRepository
	version mongodb.DocumentVersionRepository
}

func NewService(logger *slog.Logger, repo mongodb.DocumentRepository,version mongodb.DocumentVersionRepository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
		version: version,
	}
}

func (s *Service) CreateDocument(ctx context.Context, req *pb.CreateDocumentReq) (*pb.CreateDocumentRes, error) {
	s.logger.Debug("CreateDocument", "req", req)
	res, err := s.repo.CreateDocument(ctx, req)
	if err != nil {
		s.logger.Error("CreateDocument", "err", err)
		return nil, err
	}
	s.logger.Debug("CreateDocument", "res", res)
	return res, nil
}

func (s *Service) SearchDocument(ctx context.Context, req *pb.SearchDocumentReq) (*pb.SearchDocumentRes, error) {
	s.logger.Debug("SearchDocument", "req", req)
	res, err := s.repo.SearchDocument(ctx, req)
	if err != nil {
		s.logger.Error("SearchDocument", "err", err)
		return nil, err
	}
	s.logger.Debug("SearchDocument", "res", res)
	return res, nil
}

func (s *Service) GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsReq) (*pb.GetAllDocumentsRes, error) {
	s.logger.Debug("GetAllDocuments", "req", req)
	res, err := s.repo.GetAllDocuments(ctx, req)
	if err != nil {
		s.logger.Error("GetAllDocuments", "err", err)
		return nil, err
	}
	s.logger.Debug("GetAllDocuments", "res", res)
	return res, nil
}

func (s *Service) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentReq) (*pb.UpdateDocumentRes, error) {
	s.logger.Debug("UpdateDocument", "req", req)
	res, err := s.repo.UpdateDocument(ctx, req)
	if err != nil {
		s.logger.Error("UpdateDocument", "err", err)
		return nil, err
	}
	s.logger.Debug("UpdateDocument", "res", res)
	return res, nil
}

func (s *Service) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentReq) (*pb.DeleteDocumentRes, error) {
	s.logger.Debug("DeleteDocument", "req", req)
	res, err := s.repo.DeleteDocument(ctx, req)
	if err != nil {
		s.logger.Error("DeleteDocument", "err", err)
		return nil, err
	}
	s.logger.Debug("DeleteDocument", "res", res)
	return res, nil
}

func (s *Service) ShareDocument(ctx context.Context,req *pb.ShareDocumentReq) (*pb.ShareDocumentRes,error){
	res,err:=s.repo.ShareDocument(ctx,req)
	if err != nil {
		s.logger.Error("ShareDocument", "err", err)
		return nil, err
	}
	return res, nil
}

func (s *Service) GetAllVersions(ctx context.Context,req *pb.GetAllVersionsReq) (*pb.GetAllVersionsRes,error){
	res,err:=s.version.GetAllVersions(ctx,req)
	if err!=nil{
		s.logger.Error("GetAllVersions", "err", err)
		return nil, err
	}
	return &pb.GetAllVersionsRes{
		DocumentsVersion: res.Documents,
	},nil
}

func (s *Service) RestoreVersion(ctx context.Context,req *pb.RestoreVersionReq)(*pb.RestoreVersionRes,error){
	res,err:=s.version.RestoreVersion(ctx,req)
	if err!=nil{
		s.logger.Error("RestoreVersion", "err", err)
		return nil, err
	}
	return &pb.RestoreVersionRes{
		Message: res.Message,
	},nil
}