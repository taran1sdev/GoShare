package controllers

import "net/http"

type Template interface {
	Execute(http.ResponseWriter, interface{})
}
