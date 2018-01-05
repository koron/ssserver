package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/sclevine/agouti"
)

var coreArgs []string

func init() {
	coreArgs = make([]string, 0, 4)
	coreArgs = append(coreArgs,
		"headless",
		"disable-gpu",
	)
}

func setupProxy() error {
	s, ok := os.LookupEnv("HTTP_PROXY")
	if !ok {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	coreArgs = append(coreArgs, "proxy-server="+u.Host)
	if u.User != nil {
		coreArgs = append(coreArgs, "proxy-auth="+u.User.String())
	}
	return nil
}

func serve(addr string) error {
	if err := setupProxy(); err != nil {
		return err
	}

	d := agouti.ChromeDriver()
	err := d.Start()
	if err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	go func() {
		for {
			s := <-sig
			if s == os.Interrupt {
				break
			}
		}
		signal.Stop(sig)
		d.Stop()
		os.Exit(0)
	}()
	signal.Notify(sig, os.Interrupt)

	return http.ListenAndServe(addr, newHandler(d))
}

type openParams struct {
	url    string
	width  int
	height int
	wait   time.Duration
}

func newOpenParams(v url.Values) (*openParams, error) {
	p := &openParams{
		width:  1024,
		height: 768,
		wait:   0 * time.Second,
	}
	if p.url = v.Get("u"); p.url == "" {
		return nil, errors.New("u (url) must not be empty")
	}
	if s := v.Get("w"); s != "" {
		w, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("w (width) has error: %v", err)
		}
		p.width = w
	}
	if s := v.Get("h"); s != "" {
		h, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("h (height) has error: %v", err)
		}
		p.height = h
	}
	if s := v.Get("wait"); s != "" {
		wait, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("wait has error: %v", err)
		}
		p.wait = wait
	}
	return p, nil
}

func newHandler(d *agouti.WebDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := newOpenParams(r.URL.Query())
		if err != nil {
			log.Printf("parameter error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page, err := openPage(d, p)
		if err != nil {
			log.Printf("failed to open %q: %v", p.url, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer page.Destroy()
		b, err := page.Session().GetScreenshot()
		if err != nil {
			log.Printf("failed to GetScreenShot: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Disposition", "attachment")
		w.WriteHeader(http.StatusOK)
		for len(b) > 0 {
			n, err := w.Write(b)
			if err != nil {
				log.Printf("failed to Write: %v", err)
				break
			}
			b = b[n:]
		}
	}
}

func openPage(d *agouti.WebDriver, p *openParams) (*agouti.Page, error) {
	args := make([]string, 0, len(coreArgs)+4)
	args = append(args, coreArgs...)
	args = append(args, fmt.Sprintf("window-size=%d,%d", p.width, p.height))
	page, err := d.NewPage(agouti.ChromeOptions("args", args))
	if err != nil {
		return nil, err
	}
	err = page.Navigate(p.url)
	if err != nil {
		page.Destroy()
		return nil, err
	}
	if p.wait > 0 {
		time.Sleep(p.wait)
	}
	return page, nil
}

func main() {
	err := serve(":3000")
	if err != nil {
		log.Fatal("ssserve failure: %v", err)
	}
}
