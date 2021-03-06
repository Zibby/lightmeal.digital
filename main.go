package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

//RecipeInfo handles data marshlled in
type RecipeInfo struct {
	PageTitle    string   `yaml:"Title"`
	PageDesc     string   `yaml:"Description"`
	Author       string   `yaml:"Author"`
	Ingredients  []string `yaml:"Ingredients"`
	Instructions []string `yaml:"Instructions"`
	Img          string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func recipeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	rec := vars["recipe"]

	file, _ := filepath.Abs("./recipes/" + rec + "/recipe.yml")
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var recipestruct RecipeInfo

	err = yaml.Unmarshal(yamlFile, &recipestruct)
	if err != nil {
		panic(err)
	}

	tmpl := template.Must(template.ParseFiles("./static/template.html"))
	pageData := RecipeInfo{
		PageTitle:    recipestruct.PageTitle,
		PageDesc:     recipestruct.PageDesc,
		Author:       recipestruct.Author,
		Ingredients:  recipestruct.Ingredients,
		Instructions: recipestruct.Instructions,
		Img:          "/recipe-content/" + rec + "/card.jpg",
	}
	tmpl.Execute(w, pageData)
}

// RootRecipes creates cards on home page
var RootRecipes []string

type rootData struct {
	RootRecipes []string
}

func listRecipes() {
	files, err := ioutil.ReadDir("./recipes")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		RootRecipes = append(RootRecipes, f.Name())
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/index.html"))
	pageData := rootData{
		RootRecipes: RootRecipes,
	}
	tmpl.Execute(w, pageData)
}

// RequestLoggerMiddleware logs
func RequestLoggerMiddleware(r *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			defer func() {
				log.Printf(
					"[%s] [%v] %s %s %s %s %s",
					req.Method,
					time.Since(start),
					req.UserAgent(),
					req.RemoteAddr,
					req.Host,
					req.URL.Path,
					req.URL.RawQuery,
				)
			}()
			next.ServeHTTP(w, req)
		})
	}
}

func main() {
	listRecipes()
	r := mux.NewRouter()
	r.Use(RequestLoggerMiddleware(r))

	about := spaHandler{staticPath: "static", indexPath: "about.html"}
	r.Path("/").HandlerFunc(rootHandler)
	r.Path("/about").Handler(about)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/recipe/{recipe}", recipeHandler)
	r.PathPrefix("/recipe-content/").Handler(http.StripPrefix("/recipe-content/", http.FileServer(http.Dir("./recipes/"))))

	srv := &http.Server{
		Handler: r,
		Addr:    ":8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
