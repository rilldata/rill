package sharedlibrary

import (
	"embed"
	"io"
	"os"
	"path"
)

func OpenEmbed(fs embed.FS, fsPath string) (Library, error) {
	src, err := fs.Open(fsPath)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	libName := path.Base(fsPath)
	dstPath := path.Join(os.TempDir(), "sharedlibrary", libName)

	err = os.MkdirAll(path.Dir(dstPath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, err
	}

	lib, err := Open(dstPath)
	if err != nil {
		return nil, err
	}

	el := &embedLibrary{
		library: lib,
		tmpPath: dstPath,
	}

	return el, nil
}

type embedLibrary struct {
	library Library
	tmpPath string
}

func (l *embedLibrary) FindFunc(name string) (Func, error) {
	return l.library.FindFunc(name)
}

func (l *embedLibrary) Close() error {
	err := l.library.Close()
	if err != nil {
		return err
	}

	return os.Remove(l.tmpPath)
}
