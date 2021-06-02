package controllers

import (
	"fmt"
	"net/http"
	"push_service/src/models"
)

type GetEsController struct {
	BaseController
}

func (c *GetEsController) SizeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintf(w, "method not allowed!")
		return
	}

	// read request
	var search models.EsModel
	search.Init()
	res := search.Search(w, r)
	c.sendOk(w, res)
}
