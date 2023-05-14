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
	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/hilsonxhero/napoleon/cache"
	"github.com/hilsonxhero/napoleon/render"
	"github.com/hilsonxhero/napoleon/session"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

type Napoleon struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       scs.SessionManager
	DB            Database
	JetViews      jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    DatabaseConfig
	redis       RedisConfig
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

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := n.OpenDB(os.Getenv("DATABASE_TYPE"), n.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		n.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	scheduler := cron.New()
	n.Scheduler = scheduler

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = n.createClientRedisCache()
		n.Cache = myRedisCache
	}

	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = n.createClientBadgerCache()
		n.Cache = myBadgerCache

		_, err = n.Scheduler.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})

		if err != nil {
			return err
		}
	}

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
		database: DatabaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      n.BuildDSN(),
		},

		redis: RedisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	sess := session.Session{
		CookieLifetime: n.config.cookie.lifetime,
		CookiePersist:  n.config.cookie.parsist,
		CookieName:     n.config.cookie.name,
		SessionType:    n.config.sessionType,
		CookieDomain:   n.config.cookie.domain,
		DBPool:         n.DB.Pool,
	}

	switch n.config.sessionType {
	case "redis":
		sess.RedisPool = myRedisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = n.DB.Pool
	}
	n.Session = *sess.InitSession()

	n.EncryptionKey = os.Getenv("KEY")
	if n.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
		n.JetViews = *views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			// jet.InDevelopmentMode(),
		)
		n.JetViews = *views
	}

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

	if n.DB.Pool != nil {
		defer n.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}
	if badgerConn != nil {
		defer badgerConn.Close()
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
		Session:  n.Session,
	}

	n.Render = &myRenderer
}

func (n *Napoleon) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   n.createRedisPool(),
		Prefix: n.config.redis.prefix,
	}

	return &cacheClient
}

func (n *Napoleon) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: n.createBadgerConn(),
	}

	return &cacheClient
}

func (n *Napoleon) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", n.config.redis.host, redis.DialPassword(n.config.redis.password))
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (n *Napoleon) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(n.RootPath + "/tmp/badger"))

	if err != nil {
		return nil
	}

	return db
}

// BuildDSN builds the datasource name for our database, and returns it as a string
func (n *Napoleon) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		// we check to see if a database passsword has been supplied, since including "password=" with nothing
		// after it sometimes causes postgres to fail to allow a connection.
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}

	return dsn
}
