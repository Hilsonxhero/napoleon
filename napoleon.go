package napoleon

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/hilsonxhero/napoleon/render"
	"github.com/hilsonxhero/napoleon/session"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Napoleon struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	Session  scs.SessionManager
	JetViews jet.Set
	config   config
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
}

func (n *Napoleon) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middlewares"},
	}

	err := n.Init(pathConfig)
	if err != nil {
		return err
	}

	err = n.checkEnvFile(rootPath)

	if err != nil {
		return err
	}

	err = godotenv.Load(rootPath + "/.env")

	if err != nil {
		return err
	}

	infoLog, errorLog := n.startLogegrs()
	n.ErrorLog = errorLog
	n.InfoLog = infoLog
	n.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	n.Version = version
	n.RootPath = rootPath
	n.Routes = n.routes().(*chi.Mux)
	n.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			parsist:  os.Getenv("COOKIE_PARSIST"),
			srcure:   os.Getenv("SESSION_SECURE"),
			domain:   os.Getenv("SESSION_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
	}

	sess := session.Session{
		CookieLifetime: n.config.cookie.lifetime,
		CookiePersist:  n.config.cookie.parsist,
		CookieName:     n.config.cookie.name,
		SessionType:    n.config.sessionType,
		CookieDomain:   n.config.cookie.domain,
	}
	n.Session = *sess.InitSession()

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)
	n.JetViews = *views
	n.createRenderer()

	return nil
}
func (n *Napoleon) Init(p initPaths) error {
	root := p.rootPath

	for _, path := range p.folderNames {
		err := n.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Napoleon) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     n.ErrorLog,
		Handler:      n.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	n.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	n.ErrorLog.Fatal(err)
}

func (n *Napoleon) checkEnvFile(path string) error {
	err := n.CreateFileIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}

	return nil
}

func (n *Napoleon) startLogegrs() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog
}

func (n *Napoleon) createRenderer() {
	myRenderer := render.Render{
		Renderer: n.config.renderer,
		RootPath: n.RootPath,
		Port:     n.config.port,
		JetViews: n.JetViews,
	}

	n.Render = &myRenderer
}
