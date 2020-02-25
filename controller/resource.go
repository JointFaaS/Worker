package controller

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type funcResource struct {
	funcName string
	image string
	sourceCodeURL string
	sourceCodeDir string
}

func newFuncResource(funcName string, image string, sourceCodeURL string) (*funcResource, error) {
	fr := funcResource{
		funcName: funcName,
		image: image,
		sourceCodeURL: sourceCodeURL,
	}
	resp, err := http.Get(sourceCodeURL)
	if err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	z, err := os.Create(path.Join(dir, "code.zip"))
	if err != nil {
		return nil, err
	}
	io.Copy(z, resp.Body)
	fr.sourceCodeDir = dir
	err = deCompress(path.Join(dir, "code.zip"), dir)
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

func deCompress(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}

	defer reader.Close()
	for _, innerFile := range reader.File {
        info := innerFile.FileInfo()
        if info.IsDir() {
            err = os.MkdirAll(innerFile.Name, os.ModePerm)
            if err != nil {
                return err
            }
            continue
        }
        srcFile, err := innerFile.Open()
        if err != nil {
            return err
        }
        defer srcFile.Close()
        newFile, err := os.Create(path.Join(dest, innerFile.Name))
        if err != nil {
			return err
        }
        io.Copy(newFile, srcFile)
        newFile.Close()
    }
    return nil
}