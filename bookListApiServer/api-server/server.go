package bookserver

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-xorm/xorm"
	core "xorm.io/core"

	"github.com/go-macaron/auth"
	"github.com/go-macaron/binding"
	_ "github.com/lib/pq"
	"gopkg.in/macaron.v1"
)

type Book struct {
	ID     string `json:"id" xorm:"pk"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

type BookList struct {
	Items []Book `json:"items"`
}

var (
	Username = "kamol"
	Password = "hasan"
)

var Engine *xorm.Engine

func Run() {

	// DataBase management using XORM
	var err error
	connStr := "user=postgres password=kamol host=127.0.0.1 port=5432 dbname=demo sslmode=disable"
	Engine, err = xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer func() {
		Engine.Close()
		fmt.Println("engine stopped!!")
	}()

	// set logger
	f, err := os.Create("sql.log")
	if err != nil {
		panic(err)
	}
	logger := xorm.NewSimpleLogger(f)
	logger.ShowSQL(true)
	Engine.SetLogger(logger)

	// set mapping rules
	Engine.SetMapper(core.SameMapper{})

	// Create Table from structure
	err = Engine.CreateTables(Book{})
	if err != nil {
		panic(err)
	}

	// Register prometheus metrics
	Prom := RegPrometheusMetrics()

	// start server using Go-Macaron
	m := macaron.Classic()

	m.Use(macaron.Renderer(macaron.RenderOptions{}))
	m.Use(auth.Basic(Username, Password))
	m.Get("/books", GetBooks)
	m.Get("/books/:id", GetBook)
	m.Post("/books", binding.Json(BookList{}), PostBook)
	m.Post("/books/:id", binding.Json(Book{}), UpdateBook)
	m.Get("/metrics", promhttp.HandlerFor(Prom, promhttp.HandlerOpts{}))
	m.Get("/favicon.ico")
	m.NotFound(NotFoundFunc)

	log.Println("Server running... ...")
	log.Println(http.ListenAndServe(":8080", m))
}
