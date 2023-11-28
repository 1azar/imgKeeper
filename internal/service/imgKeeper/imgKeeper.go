package imgKeeper

import (
	"fmt"
	"imgKeeper/internal/lib/file"
	"imgKeeper/internal/lib/logger/sl"
	"io"
	"path/filepath"

	"context"
	imgKeeperv1 "github.com/1azar/imgKeeper-api-contracts/gen/go/imgKeeper"
	"github.com/golang/protobuf/ptypes/empty"
	"log/slog"
	"time"
)

type FileIndex interface {
	IndexFile(ctx context.Context,
		fileName string,
	) (createTime, updateTime time.Time, err error)
	GetFolder() (path string)
}

type FileProvider interface {
	GetFile(ctx context.Context, fileName string) ([]byte, error)
}

type ImgKeeper struct {
	log          *slog.Logger
	fileIndex    FileIndex
	fileProvider FileProvider
}

func New(
	log *slog.Logger,
	fileIndex FileIndex,
	fileProvider FileProvider,
) *ImgKeeper {
	return &ImgKeeper{
		log:          log,
		fileIndex:    fileIndex,
		fileProvider: fileProvider,
	}
}

func (s ImgKeeper) UploadImg(stream imgKeeperv1.ImgKeeper_UploadImgServer) error {
	const fn = "service.imgKeeper.imgKeeper.UploadImg"
	s.log.With(slog.String("fn", fn))

	myFile := file.New()
	var fileSize uint32
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		if err := myFile.OutputFile.Close(); err != nil {
			s.log.Error("failed to close file", sl.Err(err))
		}
	}()
	for {
		req, err := stream.Recv()
		if myFile.FilePath == "" {
			if err := myFile.SetFile(req.GetFileName(), s.fileIndex.GetFolder()); err != nil {
				s.log.Error("could not set file", sl.Err(err))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := req.GetChunk()
		fileSize += uint32(len(chunk))
		s.log.Debug("received a chunk with size: %d", fileSize)
		if err := myFile.Write(chunk); err != nil {
			s.log.Error("could not write chunk", sl.Err(err))
			return err
		}
	}
	fileName := filepath.Base(myFile.FilePath)
	fileCreateDate, fileUpdateDate, err := s.fileIndex.IndexFile(ctxWithTimeout, fileName)
	if err != nil {
		s.log.Error("could not index file: ", err)
		return err
	}

	s.log.Debug(fmt.Sprintf("saved file: %s, size: %d, CreateTime: %v, updateTime: %v", myFile.FilePath, fileSize, fileCreateDate, fileUpdateDate))
	return stream.SendAndClose(&imgKeeperv1.ImgUploadRes{FileName: fileName, Size: fileSize})
}

func (s ImgKeeper) DownloadImg(req *imgKeeperv1.ImgDownloadReq, steam imgKeeperv1.ImgKeeper_DownloadImgServer) error {
	//TODO implement me
	panic("implement me")
}

func (s ImgKeeper) ImgList(ctx context.Context, _ *empty.Empty) (*imgKeeperv1.ImgListRes, error) {
	//TODO implement me
	panic("implement me")
}
