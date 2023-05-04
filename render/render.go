package render

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
}

type TemplateData struct {
	isAuth     bool
	initMap    map[string]int
	StringMap  map[string]int
	FloatMap   map[string]float32
	Data       map[string]interface{}
	CSRFToken  string
	Port       string
	ServerName string
	Secure     bool
}

func (n *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(n.Renderer) {
	case "go":
		return n.GoPage(w, r, view, data)
	case "jet":
	}

	return nil
}

func (n *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", n.RootPath, view))
	if err != nil {
		return err
	}

	td := &TemplateData{}

	if data != nil {
		td = data.(*TemplateData)
	}

	err = tmpl.Execute(w, &td)

	if data != nil {
		return err
	}

	return nil
}
