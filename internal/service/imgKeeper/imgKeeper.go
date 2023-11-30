package imgKeeper

import (
	"fmt"
	"imgKeeper/internal/lib/file"
	"imgKeeper/internal/lib/logger/sl"
	"io"
	"os"
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
		filePath string,
	) (createTime, updateTime time.Time, err error)
	GetFolder() (path string)
	GetFileList(ctx context.Context) (data string, err error)
	SendToStream(stream imgKeeperv1.ImgKeeper_ImgListServer) error
}

type FileProvider interface {
	//GetFile(ctx context.Context, fileName string) (io.Reader, error)
	IsFileExist(ctx context.Context, fileName string) (ok bool, path string, err error)
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
	fileCreateDate, fileUpdateDate, err := s.fileIndex.IndexFile(ctxWithTimeout, fileName, filepath.Dir(myFile.FilePath))
	if err != nil {
		s.log.Error("could not index file: ", err)
		return err
	}

	s.log.Debug(fmt.Sprintf("saved file: %s, size: %d, CreateTime: %v, updateTime: %v", myFile.FilePath, fileSize, fileCreateDate, fileUpdateDate))
	return stream.SendAndClose(&imgKeeperv1.ImgUploadRes{FileName: fileName, Size: fileSize})
}

func (s ImgKeeper) DownloadImg(req *imgKeeperv1.ImgDownloadReq, stream imgKeeperv1.ImgKeeper_DownloadImgServer) error {
	const fn = "service.imgKeeper.imgKeeper.UploadImg"
	s.log.With(slog.String("fn", fn))

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	ok, fileLocation, err := s.fileProvider.IsFileExist(ctxWithTimeout, req.GetFileName())
	if err != nil || !ok {
		s.log.Error("could not locate file: ", err)
		return err
	}

	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	chunkSize := 1024 * 1024
	buf := make([]byte, chunkSize)
	batchNumber := 1
	for {
		num, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := buf[:num]

		//if err := stream.Send(&uploadpb.FileUploadRequest{FileName: s.filePath, Chunk: chunk}); err != nil {
		if err := stream.Send(&imgKeeperv1.ImgDownloadRes{FileName: filepath.Base(fileLocation), Chunk: chunk}); err != nil {
			return err
		}
		s.log.Debug(fmt.Sprintf("Sent - batch #%v - size - %v", batchNumber, len(chunk)))
		batchNumber += 1
	}
	s.log.Debug(fmt.Sprintf("file %s sent", fileLocation))
	return nil
}

func (s ImgKeeper) ImgList(_ *empty.Empty, stream imgKeeperv1.ImgKeeper_ImgListServer) error {
	const fn = "service.imgKeeper.imgKeeper.ImgList"
	s.log.With(slog.String("fn", fn))

	if err := s.fileIndex.SendToStream(stream); err != nil {
		s.log.Error("could not send file list to client", err)
		return err
	}

	s.log.Debug("file list sent to client")
	return nil
}
