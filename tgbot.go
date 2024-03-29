package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/zhao-kun/reminder-tgbot/model"
	"github.com/zhao-kun/reminder-tgbot/repo"
	"github.com/zhao-kun/reminder-tgbot/server"
	"github.com/zhao-kun/reminder-tgbot/telegram"
)

type (
	response struct {
		Ok bool `json:"ok"`
	}
)

// wrapClientRepo wrap a func with config parameter
func wrapClientRepo(c telegram.Client, r repo.Repo,
	f func(telegram.Client, repo.Repo, model.TgMessage)) rest.HandlerFunc {
	return func(w rest.ResponseWriter, req *rest.Request) {
		ok := func() {
			w.WriteJson(response{true})
			w.WriteHeader(http.StatusOK)
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			rest.Error(w, "read request body error", http.StatusBadGateway)
			return
		}

		log.Printf("Request %s is comming body is:\n%s\n", req.URL.Path[1:], body)
		var message model.TgMessage
		err = json.Unmarshal(body, &message)
		if err != nil {
			log.Printf("Can't unmarshal body [%s] to message", body)
			ok()
			return
		}
		f(c, r, message)
		ok()
	}

}
func startServer(c telegram.Client, r repo.Repo) (<-chan error, error) {

	checkInHandle := wrapClientRepo(c, r, server.TelegramServerHandle)
	router, err := rest.MakeRouter(
		rest.Post(r.Cfg().WebhookEndpoint, checkInHandle),
		rest.Put(r.Cfg().WebhookEndpoint, checkInHandle),
	)
	if err != nil {
		log.Fatalf("Make router error :%s", err)
		return nil, err
	}

	apiServer := rest.NewApi()
	apiServer.Use(rest.DefaultCommonStack...)
	apiServer.SetApp(router)

	done := make(chan error, 1)
	go func() {
		server := &http.Server{
			Addr:    r.Cfg().ListenAddr,
			Handler: apiServer.MakeHandler(),
		}

		log.Printf("Start Listen on: %s", r.Cfg().ListenAddr)
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("ListenAndServe on %s error:%s", r.Cfg().ListenAddr, err)
			done <- err
		}
	}()
	return done, nil
}

func readConf(path string) (cfg model.Config, err error) {
	cfg.ListenAddr = ":8888"
	cfg.CheckUesrs = []string{"zhaokun"}
	// a chinese festival calendar service maybe broken in feature
	cfg.CNCalendarServiceEndpoint = "http://api.goseek.cn/Tools/holiday"
	file, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return cfg, fmt.Errorf("open file %s error %s", path, err)
	}

	c, err := ioutil.ReadAll(file)
	if err != nil {
		return cfg, fmt.Errorf("read file contents error %s", err)
	}

	err = json.Unmarshal(c, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("unmarshal [%s] error", string(c))
	}

	if cfg.WebhookEndpoint == "" {
		return cfg, fmt.Errorf("WebhookEndpoint is required")
	}

	return
}

func main() {
	bdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	config, err := readConf(fmt.Sprintf("%s/config.json", bdir))
	if err != nil {
		log.Fatalf("readConf error %s", err)
	}

	r := repo.New(config)
	c := telegram.NewClient(config)

	err = server.StartAllBotTask(c, r)
	if err != nil {
		log.Fatalf("%s", err)
	}

	done, err := startServer(c, r)
	if err != nil {
		log.Fatalf("boot server error")
		return
	}

	for {
		select {
		case err = <-done:
			if err != nil {
				log.Fatalf("start server error %s, exit", err)
			}
		}
	}
}
