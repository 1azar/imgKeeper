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
	imgKeeper   ImgKeeper
	FileLimiter chan struct{} // ограничитель на скачивание/загрузку
	ListLimiter chan struct{} // ограничитель на просмотр списка файлов
}

func Register(gRPCServer *grpc.Server, imgKeeper ImgKeeper) {
	imgKeeperv1.RegisterImgKeeperServer(gRPCServer, &serverAPI{
		imgKeeper:   imgKeeper,
		FileLimiter: make(chan struct{}, 10),
		ListLimiter: make(chan struct{}, 100),
	})
}

func (s *serverAPI) UploadImg(stream imgKeeperv1.ImgKeeper_UploadImgServer) error {
	s.FileLimiter <- struct{}{} //если канал заполнени функция лочится пока не освободится место
	if err := s.imgKeeper.UploadImg(stream); err != nil {
		<-s.FileLimiter // освобождаем очередь
		return err
	}
	<-s.FileLimiter // освобождаем очередь
	return nil
}

func (s *serverAPI) DownloadImg(req *imgKeeperv1.ImgDownloadReq, stream imgKeeperv1.ImgKeeper_DownloadImgServer) error {
	s.FileLimiter <- struct{}{}
	if err := s.imgKeeper.DownloadImg(req, stream); err != nil {
		<-s.FileLimiter
		return err
	}
	<-s.FileLimiter
	return nil
}

func (s *serverAPI) ImgList(ctx context.Context, _ *empty.Empty) (*imgKeeperv1.ImgListRes, error) {
	panic("implement me")
}
