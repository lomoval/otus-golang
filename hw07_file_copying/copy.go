package main

import (
	"errors"
	"io"
	"os"
)

var (
	ErrOpenFile              = errors.New("open file error")
	ErrIncorrectOffset       = errors.New("incorrect offset")
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrCopying               = errors.New("failed to copy file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

const maxCopyLen = int64(1024)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if fromPath == toPath {
		return ErrCopying
	}
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return ErrOpenFile
	}
	defer fromFile.Close()

	stat, err := fromFile.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}

	if offset > stat.Size() {
		return ErrOffsetExceedsFileSize
	}

	if offset < 0 {
		offset = 0
	}

	copySize := stat.Size() - offset
	if limit > 0 && limit < copySize {
		copySize = limit
	}

	_, err = fromFile.Seek(offset, io.SeekStart)
	if err != nil {
		return ErrIncorrectOffset
	}

	toFile, err := os.Create(toPath)
	if err != nil {
		return ErrOpenFile
	}
	defer toFile.Close()

	progressChan := make(chan int64)
	terminatedPb := progressBar{}.start(progressChan, copySize)
	defer func() {
		close(progressChan)
		<-terminatedPb
	}()

	copiedBytes := int64(0)
	copyLen := maxCopyLen
	for copiedBytes < copySize {
		if copyLen+copiedBytes > copySize {
			copyLen = copySize - copiedBytes
		}
		_, err := io.CopyN(toFile, fromFile, copyLen)
		copiedBytes += copyLen
		progressChan <- copiedBytes
		if err != nil && !errors.Is(err, io.EOF) {
			return ErrCopying
		}
	}
	return nil
}
