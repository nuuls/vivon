package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/sirupsen/logrus"
)

const redirURL = "https://api.twitch.tv/kraken/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=chat_login user_read&state=%s"

type config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirURL     string `json:"redirUrl"`
	Addr         string `json:"addr"`
}

var log = logrus.StandardLogger()
var cfg config

var client = &http.Client{
	Timeout: time.Second * 15,
}

func main() {
	configPath := flag.String("config", "./config/dev.json", "path to config file")
	flag.Parse()
	cfg = loadConfig(*configPath)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.HandleFunc("/login", login)
	r.HandleFunc("/auth", auth)
	r.Mount("/", http.StripPrefix("/", http.FileServer(http.Dir("../app"))))
	srv := &http.Server{
		ReadTimeout: time.Second * 30,
		Handler:     r,
		Addr:        cfg.Addr,
	}
	log.WithField("addr", cfg.Addr).Info("listening")
	log.Fatal(srv.ListenAndServe())
}

func auth(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "")
		return
	}
	values := url.Values{}
	values.Add("client_id", cfg.ClientID)
	values.Add("client_secret", cfg.ClientSecret)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", cfg.RedirURL)
	values.Add("code", code)
	req, err := http.NewRequest(http.MethodPost,
		"https://api.twitch.tv/kraken/oauth2/token", strings.NewReader(values.Encode()))
	if err != nil {
		log.WithError(err)
		writeError(w, 500, "")
		return
	}
	res, err := client.Do(req)
	if err != nil {
		log.WithError(err)
		writeError(w, 500, "")
		return
	}
	data := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = unmarshalJSON(res.Body, &data)
	if err != nil {
		log.WithError(err)
		writeError(w, 500, "")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "twitch_token",
		Value: data.AccessToken,
	})
	http.Redirect(w, r, "/zulul.html", 307)
}

func login(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf(redirURL, cfg.ClientID, cfg.RedirURL, "kappa"), 307)
}

func loadConfig(path string) config {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	cfg := config{}
	err = json.Unmarshal(bs, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
