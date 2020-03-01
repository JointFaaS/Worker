package controller

import (
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
	resp, err := http.Get(sourceCodeURL)
	if err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	z, err := os.Create(path.Join(dir, "source"))
	if err != nil {
		return nil, err
	}
	io.Copy(z, resp.Body)
	fr := funcResource{
		funcName: funcName,
		image: image,
		sourceCodeURL: sourceCodeURL,
		sourceCodeDir: dir,
	}
	return &fr, nil
}
