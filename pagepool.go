package main

import (
	"errors"
	"log"
	"sync"

	"github.com/sclevine/agouti"
)

// Closed is the error returned by Get when the PagePool is closed already.
var Closed = errors.New("page pool is closed")

type pageEntry struct {
	u bool
	p *agouti.Page
}

// PagePool is a pool for *agouti.Page
type PagePool struct {
	drv   *agouti.WebDriver
	max   int
	c     *sync.Cond
	pages []*pageEntry
	err   error
}

// NewPool creates a page pool with driver.
func NewPool(drv *agouti.WebDriver, max int) *PagePool {
	pp := &PagePool{
		drv:   drv,
		max:   max,
		c:     sync.NewCond(&sync.Mutex{}),
		pages: make([]*pageEntry, 0, 4),
	}
	return pp
}

// Get returns a page can be used.  After finished to use, return with Put
// method.
func (pp *PagePool) Get() (*agouti.Page, error) {
	if pp.err != nil {
		return nil, pp.err
	}
	pp.c.L.Lock()
	for pp.activePage() == pp.max {
		pp.c.Wait()
	}
	defer pp.c.L.Unlock()

	pe, n := pp.freePage()
	if pe != nil {
		pe.u = true
		log.Printf("PagePool.Get: cached #%d", n)
		return pe.p, nil
	}
	p, err := pp.drv.NewPage()
	if err != nil {
		return nil, err
	}
	log.Printf("PagePool.Get: allocated #%d", len(pp.pages))
	pe = &pageEntry{p: p, u: true}
	pp.pages = append(pp.pages, pe)
	return pe.p, nil
}

// Put releases back a page to the pool.
func (pp *PagePool) Put(p *agouti.Page) {
	pp.c.L.Lock()
	defer pp.c.L.Unlock()

	for i, pe := range pp.pages {
		if pe.p == p {
			pe.u = false
			log.Printf("PagePool.Put: released #%d", i)
			pp.c.Broadcast()
			return
		}
	}

	log.Print("PagePool.Put: unmanaged page")
}

func (pp *PagePool) freePage() (*pageEntry, int) {
	for i, pe := range pp.pages {
		if !pe.u {
			return pe, i
		}
	}
	return nil, -1
}

func (pp *PagePool) activePage() int {
	n := 0
	for _, p := range pp.pages {
		if p.u {
			n++
		}
	}
	return n
}

// Close closes all pages and finish the pool.
func (pp *PagePool) Close() {
	log.Printf("PagePool.Close: waiting")
	pp.c.L.Lock()
	pp.err = Closed
	for pp.activePage() != 0 {
		pp.c.Wait()
	}
	defer pp.c.L.Unlock()
	log.Printf("PagePool.Close: closing")

	for _, pe := range pp.pages {
		pe.p.Destroy()
	}
	pp.pages = nil
	log.Printf("PagePool.Close: closed")
}
