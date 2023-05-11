package napoleon

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
)

func (n *Napoleon) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// WriteXML writes xml from arbitrary data
func (c *Napoleon) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// DownloadFile downloads a file
func (c *Napoleon) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

// Error404 returns page not found response
func (c *Napoleon) Error404(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusNotFound)
}

// Error500 returns internal server error response
func (c *Napoleon) Error500(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusInternalServerError)
}

// ErrorUnauthorized sends an unauthorized status (client is not known)
func (c *Napoleon) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusUnauthorized)
}

// ErrorForbidden returns a forbidden status message (client is known)
func (c *Napoleon) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusForbidden)
}

// ErrorStatus returns a response with the supplied http status
func (c *Napoleon) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
