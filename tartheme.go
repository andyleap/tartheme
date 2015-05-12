// tartheme project tartheme.go
package tartheme

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	mmapgo "github.com/edsrzf/mmap-go"
)

type TarTheme struct {
	Assets
	tar mmapgo.MMap
}

type Asset struct {
	Name    string
	Data    []byte
	ModTime time.Time
}

type Assets map[string]*Asset

func Load(file string) (*TarTheme, error) {
	tt := &TarTheme{}
	tt.Assets = make(Assets)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	tt.tar, err = mmapgo.Map(f, mmapgo.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	tt.readAllAssets()

	return tt, nil
}

func LoadDir(path string) (*TarTheme, error) {
	tt := &TarTheme{}
	tt.Assets = make(Assets)

	filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		data, _ := ioutil.ReadFile(file)
		if info.IsDir() {
			return nil
		}
		asset := &Asset{}
		asset.Name = filepath.ToSlash(file[len(path):])
		if asset.Name[0] == '/' {
			asset.Name = asset.Name[1:]
		}
		asset.Data = data
		asset.ModTime = info.ModTime()
		tt.Assets[asset.Name] = asset
		return nil
	})

	return tt, nil
}

func (a Assets) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	upath := req.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = upath[1:]
		req.URL.Path = upath
	}
	asset, ok := a[upath]
	if !ok {
		rw.WriteHeader(404)
		return
	}
	http.ServeContent(rw, req, path.Base(req.URL.Path), asset.ModTime, bytes.NewReader(asset.Data))
}

func (a Assets) Prefix(prefix string) Assets {
	na := make(Assets)
	for name, asset := range a {
		if strings.HasPrefix(name, prefix) {
			na[name[len(prefix):]] = asset
		}
	}
	return na
}

func (a Assets) Templates() *template.Template {
	t := template.New("")

	return a.AddTemplates(t)
}

func (a Assets) AddTemplates(t *template.Template) *template.Template {
	for name, asset := range a {
		t.New(name).Parse(string(asset.Data))
	}

	return t
}
