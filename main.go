package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func chanWriter(opChan chan string) {
	for index := 0; index < 5; index++ {
		opChan <- fmt.Sprintf("num: %d", index)
		time.Sleep(time.Second)
	}
	close(opChan)
}

func ginServer(addr string) {
	api := gin.Default()
	api.GET("/:id", func(c *gin.Context) {
		taskID := c.Param("id")
		opChan := make(chan string)
		go chanWriter(opChan)

		c.Stream(func(w io.Writer) bool {
			output, ok := <-opChan
			if !ok {
				return false
			}
			outputBytes := bytes.NewBufferString(taskID + output)
			c.Writer.Write(append(outputBytes.Bytes(), []byte("\n")...))
			return true
		})
	})

	api.Run(addr)
}

func httpServer(addr string) {
	srv := http.Server{
		Addr: addr,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		opChan := make(chan string)
		go chanWriter(opChan)
		w.WriteHeader(200)
		for output := range opChan {
			outputBytes := bytes.NewBufferString(output)
			w.Write(append(outputBytes.Bytes(), []byte("\n")...))
			w.(http.Flusher).Flush()
		}
	})

	srv.ListenAndServe()
}

func main() {
	go ginServer(":8081")
	go httpServer(":8082")

	<-make(chan struct{})
	return
}