package main

import (
	"errors"
	"io"
	"os"

	//nolint:depguard
	"github.com/cheggaaa/pb/v3" // Progress Bar
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	srcFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	srcFileSize := srcFileInfo.Size()
	// fmt.Println("srcFileSize:", srcFileSize) // См.: du -bs /имя/исходного/файла

	// Является ли файл обычным файлом (не символьной ссылкой, не директорией и т.д.)
	if !srcFileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	// offset больше, чем размер файла - невалидная ситуация
	if offset > srcFileSize {
		return ErrOffsetExceedsFileSize
	}

	// fmt.Println("io.SeekStart:", io.SeekStart) // 0, константа
	if _, err := srcFile.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	copyBytes := limit // Копируем limit байт, но есть нюансы
	if limit == 0 || offset+limit > srcFileSize {
		// 1) Если limit равен нулю, считаем, что копируем файл до конца (0 - дефолтное значение для int64)
		// 2) limit больше, чем размер файла - валидная ситуация, копируется исходный файл до его EOF
		copyBytes = srcFileSize - offset
	}
	if copyBytes == 0 {
		// При этом если нам нужно по итогу этой проверки скопировать 0 байт
		// (т.е. если srcFileSize равно offset или исходный файл пустой),
		// то просто возвращаем nil (создали пустой файл, но ничего в него не скопировали)
		return nil
	}

	copyPB := pb.Full.Start64(copyBytes)
	copyPBReader := copyPB.NewProxyReader(srcFile)

	_, err = io.CopyN(destFile, copyPBReader, copyBytes)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	copyPB.Finish()
	return nil
}
