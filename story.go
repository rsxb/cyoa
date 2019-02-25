package cyoa

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

// default HTML template
var tmpl *template.Template

const defaultHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{ .Title }} - Choose Your Own Adventure</title>
    <style>
      body {
        font-family: helvetica, arial;
      }
      h1 {
        text-align: center;
        position: relative;
      }
      .page {
        width: 80%;
        max-width: 500px;
        margin: auto;
        margin-top: 40px;
        margin-bottom: 40px;
        padding: 80px;
        background: #fffcf6;
        border: 1px solid #eee;
        box-shadow: 0 10px 6px -6px #777;
      }
      ul {
        border-top: 1px dotted #ccc;
        padding: 10px 0 0 0;
      }
      li {
        padding-top: 10px;
      }
      a,
      a:visited {
        text-decoration: none;
        color: #6295b5;
      }
      a:active,
      a:hover {
        color: #7792a2;
      }
      p {
        text-indent: 1em;
      }
    </style>
  </head>
  <body>
    <section class="page">
      <h1>{{ .Title }}</h1>
      {{ range .Paragraphs }}
      <p>{{.}}</p>
      {{ end }}
      <ul>
        {{ range .Options }}
        <li>
          <a href="/{{ .Chapter }}">{{ .Text }}</a>
        </li>
        {{ end }}
      </ul>
    </section>
  </body>
</html>`

func init() {
	tmpl = template.Must(template.New("").Parse(defaultHTML))
}

// parsePath gets the chapter title from a request's URL path.
func parsePath(r *http.Request) string {
	path := r.URL.Path
	if path == "" || path == "/" {
		path = "/intro"
	}
	return path[1:]
}

type handler struct {
	story     Story
	template  *template.Template
	parsePath func(r *http.Request) string
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.parsePath(r)
	if chapter, ok := h.story[path]; ok {
		err := h.template.Execute(w, chapter)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Chapter not found.", http.StatusNotFound)
}

// HandlerOption configures Story handlers.
type HandlerOption func(h *handler)

// WithTemplate applies an html.Template to the returned handler.
func WithTemplate(t *template.Template) HandlerOption {
	return func(h *handler) {
		h.template = t
	}
}

// WithParser applies a custom URL parser to the returned handler.
func WithParser(pathParser func(r *http.Request) string) HandlerOption {
	return func(h *handler) {
		h.parsePath = pathParser
	}
}

// NewHandler returns an http.Handler that parses story templates.
func NewHandler(s Story, opts ...HandlerOption) http.Handler {
	h := handler{s, tmpl, parsePath}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

// Story is a Choose-Your-Own-Adventure plotline.
type Story map[string]Chapter

// Chapter is a section of a story.
type Chapter struct {
	Title      string   `json:"title,omitempty"`
	Paragraphs []string `json:"story,omitempty"`
	Options    []Option `json:"options,omitempty"`
}

// Option is a choice presented to the user.
type Option struct {
	Text    string `json:"text,omitempty"`
	Chapter string `json:"arc,omitempty"`
}

// FromJSON converts from JSON to Story.
func FromJSON(r io.Reader) (Story, error) {
	var story Story
	d := json.NewDecoder(r)
	if err := d.Decode(&story); err != nil {
		return nil, fmt.Errorf("FromJSON: %s", err)
	}
	return story, nil
}
