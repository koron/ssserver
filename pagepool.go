package main

import (
	"errors"
	"log"
	"sync"

	"github.com/sclevine/agouti"
	"golang.org/x/sync/semaphore"
)

// Closed is the error returned by Get when the PagePool is closed already.
var Closed = errors.New("page pool is closed")

// PagePool is a pool for *agouti.Page
type PagePool struct {
	drv   *agouti.WebDriver
	max   int64
	c     *sync.Cond
	sem   *semaphore.Weighted
	pool  *sync.Pool
	pages []*agouti.Page
	err   error
}

// NewPool creates a page pool with driver.
func NewPool(drv *agouti.WebDriver, max int) *PagePool {
	pp := &PagePool{
		drv: drv,
		max: int64(max),
		c:   sync.NewCond(&sync.Mutex{}),
		sem: semaphore.NewWeighted(int64(max)),
	}
	pp.pool = &sync.Pool{New: pp.newPage}
	return pp
}

type newPage struct {
	p   *agouti.Page
	err error
}

func (pp *PagePool) newPage() interface{} {
	page, err := pp.drv.NewPage()
	if err != nil {
		return &newPage{err: err}
	}
	pp.pages = append(pp.pages, page)
	log.Printf("page allocated: %d", len(pp.pages))
	return &newPage{p: page}
}

// Get returns a page can be used.  After finished to use, return with Put
// method.
func (pp *PagePool) Get() (*agouti.Page, error) {
	if pp.err != nil {
		return nil, pp.err
	}
	pp.c.L.Lock()
	for !pp.sem.TryAcquire(1) {
		pp.c.Wait()
	}
	defer pp.c.L.Unlock()

	r := pp.pool.Get().(*newPage)
	if r.err != nil {
		return nil, r.err
	}
	log.Printf("current pages: %d", len(pp.pages))
	return r.p, nil
}

// Put releases back a page to the pool.
func (pp *PagePool) Put(p *agouti.Page) {
	pp.c.L.Lock()
	defer pp.c.L.Unlock()

	pp.pool.Put(&newPage{p: p})
	pp.sem.Release(1)
	pp.c.Broadcast()
	log.Printf("page is returned")
}

// Close closes all pages and finish the pool.
func (pp *PagePool) Close() {
	pp.c.L.Lock()
	pp.err = Closed
	for !pp.sem.TryAcquire(pp.max) {
		pp.c.Wait()
	}
	defer pp.c.L.Unlock()

	for _, p := range pp.pages {
		p.Destroy()
	}
	pp.pages = nil
	pp.sem.Release(pp.max)
}
