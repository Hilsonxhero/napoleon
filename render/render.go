package render

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
	JetViews   jet.Set
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
		return n.JetPage(w, r, view, variables, data)
	default:
	}

	return errors.New("no rendering engine specified")
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
func (n *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	t, err := n.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))

	if err != nil {
		log.Println(err)
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
