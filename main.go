package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
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
		"hide-scrollbars",
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

	drv := agouti.ChromeDriver()
	err := drv.Start()
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
		drv.Stop()
		os.Exit(0)
	}()
	signal.Notify(sig, os.Interrupt)

	return http.ListenAndServe(addr, newHandler(drv))
}

type openParams struct {
	url     string
	width   int
	height  int
	scrollX int
	scrollY int
	wait    time.Duration
	full    bool
	save    bool
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
	if s := v.Get("sX"); s != "" {
		sx, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("sX (scrollX) has error: %v", err)
		}
		p.scrollX = sx
	}
	if s := v.Get("sY"); s != "" {
		sy, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("sY (scrollY) has error: %v", err)
		}
		p.scrollY = sy
	}
	if s := v.Get("wait"); s != "" {
		wait, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("wait has error: %v", err)
		}
		p.wait = wait
	}
	if _, ok := v["full"]; ok {
		p.full = true
	}
	if _, ok := v["save"]; ok {
		p.save = true
	}
	return p, nil
}

func scrollTo(page *agouti.Page, x, y int) error {
	if x == 0 && y == 0 {
		return nil
	}
	return page.RunScript("window.scrollTo(x, y)", map[string]interface{}{
		"x": x,
		"y": y,
	}, nil)
}

func getScreenshot(page *agouti.Page, full bool) ([]byte, error) {
	if !full {
		return page.Session().GetScreenshot()
	}
	var v []int
	err := page.RunScript(`return [window.innerWidth, window.innerHeight, window.document.body.clientHeight]`, nil, &v)
	if err != nil {
		return nil, err
	}
	w, dy, h := v[0], v[1], v[2]
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y += dy {
		if y+dy > h {
			y = h - dy
		}
		err := scrollTo(page, 0, y)
		if err != nil {
			return nil, err
		}
		b, err := page.Session().GetScreenshot()
		if err != nil {
			return nil, err
		}
		// concatenate screenshots.
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		draw.Draw(dst, image.Rect(0, y, w, y+dy), img, image.Pt(0, 0), draw.Src)
	}
	bb := &bytes.Buffer{}
	err = png.Encode(bb, dst)
	if err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func newHandler(drv *agouti.WebDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := newOpenParams(r.URL.Query())
		if err != nil {
			log.Printf("parameter error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page, err := openPage(drv, p)
		if err != nil {
			log.Printf("failed to open %q: %v", p.url, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer page.Destroy()
		b, err := getScreenshot(page, p.full)
		if err != nil {
			log.Printf("failed to get screenshot: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if p.save {
			w.Header().Add("Content-Disposition", "attachment")
		} else {
			w.Header().Add("Content-Disposition", "inline")
		}
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

func openPage(drv *agouti.WebDriver, p *openParams) (*agouti.Page, error) {
	args := make([]string, 0, len(coreArgs)+4)
	args = append(args, coreArgs...)
	args = append(args, fmt.Sprintf("window-size=%d,%d", p.width, p.height))
	page, err := drv.NewPage(agouti.ChromeOptions("args", args))
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
	if !p.full {
		scrollTo(page, p.scrollX, p.scrollY)
	}
	return page, nil
}

func main() {
	var (
		addr string
	)
	flag.StringVar(&addr, "addr", ":3000", "server listen address")
	flag.Parse()
	err := serve(addr)
	if err != nil {
		log.Fatal("ssserve failure: %v", err)
	}
}
