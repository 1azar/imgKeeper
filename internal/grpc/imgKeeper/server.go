package imgKeeper

import (
	"context"
	imgKeeperv1 "github.com/1azar/imgKeeper-api-contracts/gen/go/imgKeeper"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type ImgKeeper interface {
	UploadImg(stream imgKeeperv1.ImgKeeper_UploadImgServer) error
	DownloadImg(req *imgKeeperv1.ImgDownloadReq, steam imgKeeperv1.ImgKeeper_DownloadImgServer) error
	ImgList(ctx context.Context, _ *empty.Empty) (*imgKeeperv1.ImgListRes, error)
}

type serverAPI struct {
	imgKeeperv1.UnimplementedImgKeeperServer
	imgKeeper ImgKeeper
}

func Register(gRPCServer *grpc.Server, imgKeeper ImgKeeper) {
	imgKeeperv1.RegisterImgKeeperServer(gRPCServer, &serverAPI{imgKeeper: imgKeeper})
}

func (s *serverAPI) UploadImg(stream imgKeeperv1.ImgKeeper_UploadImgServer) error {
	if err := s.imgKeeper.UploadImg(stream); err != nil {
		return err
	}
	return nil
}

func (s *serverAPI) DownloadImg(req *imgKeeperv1.ImgDownloadReq, steam imgKeeperv1.ImgKeeper_DownloadImgServer) error {
	panic("implement me")
}

func (s *serverAPI) ImgList(ctx context.Context, _ *empty.Empty) (*imgKeeperv1.ImgListRes, error) {
	panic("implement me")
}
