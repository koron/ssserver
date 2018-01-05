package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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

func newHandler(d *agouti.WebDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Path
		for u[0] == '/' {
			u = u[1:]
		}
		sw := 1024
		sh := 768
		dur := 0 * time.Second
		p, err := openPage(d, u, sw, sh)
		if err != nil {
			log.Printf("failed to open %q: %v", u, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer p.Destroy()
		if dur > 0 {
			time.Sleep(dur)
		}
		b, err := p.Session().GetScreenshot()
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

func openPage(d *agouti.WebDriver, uri string, w, h int) (*agouti.Page, error) {
	args := make([]string, 0, len(coreArgs)+4)
	args = append(args, coreArgs...)
	args = append(args, fmt.Sprintf("window-size=%d,%d", w, h))
	p, err := d.NewPage(agouti.ChromeOptions("args", args))
	if err != nil {
		return nil, err
	}
	err = p.Navigate(uri)
	if err != nil {
		p.Destroy()
		return nil, err
	}
	return p, nil
}

func main() {
	err := serve(":3000")
	if err != nil {
		log.Fatal("ssserve failure: %v", err)
	}
}
