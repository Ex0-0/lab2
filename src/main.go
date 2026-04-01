package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"simple-web-server/src/client"
	"simple-web-server/src/tracer"
	"sync"
)

type TemplateHandler struct {
	once     sync.Once
	template *template.Template
	filename string
}

func (t *TemplateHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	t.once.Do(func() {
		t.template = template.Must(template.ParseFiles(filepath.Join("src", "templates", t.filename)))
	})

	t.template.Execute(writer, req)
}

func main() {
	var addr = flag.String("addr", ":8080", "Адрес для запуска сервера")
	flag.Parse()

	r := client.NewRoom()
	r.Tracer = tracer.New(os.Stdout)

	http.Handle("/", &TemplateHandler{filename: "index.html"})
	http.Handle("/room", r)

	go r.Run()

	fmt.Printf("🌐 HTTP сервер запущен на http://localhost%s\n", *addr)
	fmt.Println("💡 Нажмите Ctrl+C для остановки сервера")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Ошибка запуска сервера: %v\n", err)
	}
}