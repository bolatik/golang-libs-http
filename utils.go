package http

import (
	"encoding/json"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"errors"
	"net/http"
	"strconv"

	"github.com/ajg/form"
)

var (
	ErrBadRequest = errors.New("last and first params are using together")
)

func ParseListParams(r *http.Request) (*ListParams, error) {
	var ipp = 10
	var err error
	ippStr := r.URL.Query().Get("ipp")
	if ippStr != "" {
		ipp, err = strconv.Atoi(ippStr)
		if err != nil {
			ipp = 10
		}
	}

	var seq = &Sequence{SequenceNone, 0}
	first := r.URL.Query().Get("first")
	if first != "" {
		qty, err := strconv.Atoi(first)
		if err != nil {
			return nil, err
		}
		seq = &Sequence{SequenceFirst, qty}
	}
	last := r.URL.Query().Get("last")
	if last != "" {
		qty, err := strconv.Atoi(last)
		if err != nil {
			return nil, err
		}
		seq = &Sequence{SequenceLast, qty}
	}
	if last != "" && first != "" {
		return nil, ErrBadRequest
	}

	var p = 1
	pStr := r.URL.Query().Get("p")
	if pStr != "" {
		p, err = strconv.Atoi(pStr)
		if err != nil {
			p = 1
		}
	}

	var fields []string
	cf := r.URL.Query().Get("fields")
	if cf != "" {
		fields = strings.Split(cf, ",")
	}

	queryMap, err := ParseJsonParam(r, "q")
	if err != nil {
		return nil, err
	}

	sortStr := r.URL.Query().Get("sort")
	order := true
	if sortStr != "" && sortStr[0] == '-' {
		order = false
		sortStr = sortStr[1:]
	}

	lparams := &ListParams{}
	lparams.Pagination.ItemsPerPage = ipp
	lparams.Seq = seq
	lparams.Pagination.Page = p
	lparams.Fields = fields
	lparams.Query = queryMap

	if sortStr != "" {
		lparams.Sort = map[string]bool{
			sortStr: order,
		}
	}

	return lparams, nil
}

func ParseBody(r *http.Request, item interface{}) error {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		return ParseJSON(r, item)
	case "application/x-www-form-urlencoded":
		return ParseForm(r, item)
	}
	return ParseJSON(r, item)
}

func ParseJSON(r *http.Request, item interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(item)
	if err != nil {
		return err
	}

	return nil
}

func ParseForm(r *http.Request, obj interface{}) error {
	err := form.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func ParseJsonParam(r *http.Request, paramName string) (map[string]interface{}, error) {
	paramMap := bson.M{}
	qs := r.URL.Query().Get(paramName)
	if qs == "" {
		paramMap = nil
	} else {
		if err := json.Unmarshal([]byte(qs), &paramMap); err != nil {
			return nil, err
		}
	}

	return paramMap, nil
}

func OK(d interface{}) Response {
	resp := &response{}
	resp.Status = http.StatusOK
	resp.Data = d
	return resp
}

func Created(d interface{}) Response {
	resp := &response{}
	resp.Status = http.StatusCreated
	resp.Data = d
	return resp
}
