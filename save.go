package gocrud

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/vardius/gorepo"
	"github.com/vardius/goserver"
	"golang.org/x/net/context"
)

type saveAction struct{}

func (act *saveAction) Handle(ctx context.Context, w http.ResponseWriter, req *http.Request, c *goserver.Context, repo gorepo.Repository, t reflect.Type) {
	var err error
	data := reflect.New(reflect.SliceOf(t)).Interface()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = req.Body.Close(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, data); err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	err = doSave(ctx, repo, data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	} else if err = json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func doSave(ctx context.Context, repo gorepo.Repository, data interface{}) error {
	c := make(chan error, 1)
	go func() {
		c <- repo.Save(data)
	}()
	select {
	case <-ctx.Done():
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func init() {
	Register("save", &saveAction{})
}
