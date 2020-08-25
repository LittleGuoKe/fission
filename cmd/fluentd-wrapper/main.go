package main

import (
	"fmt"
	"github.com/fission/fission/cmd/fluentd-wrapper/wrapper"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func updateFluentd(w http.ResponseWriter, req *http.Request) {
	// jingtao add do something
	fmt.Print("updateFluentd\n")
	wrapper.ReloadFluentd()
	w.Write([]byte("success"))

}

// jingtao add 不是必须选项，暂时不实现
func checkFluentd(w http.ResponseWriter, req *http.Request) {
	fmt.Print("checkFluentd!")
	w.Write([]byte("not implement"))
}

func main() {
	// start fluentd
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	wrapper.StartFluentd(zapLogger)
	http.HandleFunc("/update", updateFluentd)
	http.HandleFunc("/check", checkFluentd)
	http.ListenAndServe(":8090", nil)
}
