package crawler

import (
	"github.com/angrymuskrat/instagram-auditor/crawler/data"
	"github.com/visheratin/unilog"
	"go.uber.org/zap"
	"time"
)

type entity struct {
	id string
	err error
	profile *data.Profile
}

type Crawler struct {
	workers []*worker
	inCh chan entity
	outCh chan entity
}

func New(ports []int) *Crawler {
	cr := Crawler{
		inCh:        make(chan entity),
		outCh:       make(chan entity),
	}
	cr.workers = make([]*worker, len(ports))
	for i, p := range ports {
		cr.workers[i] = &worker{
			id:         i,
			inCh:       cr.inCh,
			outCh:      cr.outCh,
		}
		cr.workers[i].init(p)
		go cr.workers[i].start()
	}
	return &cr
}

func (c *Crawler) Start(ids []string) {
	go c.distributeEntities(ids)

	for e := range c.outCh {
		if e.err == nil {
			unilog.Logger().Info("success collect", zap.String("nickname", e.profile.Username))
		}
	}
}

func (c *Crawler) distributeEntities(ids []string) {
	for _, id := range ids {
		c.inCh <- entity{id:id}
	}
	time.Sleep(1 * time.Second)
	close(c.inCh)
	close(c.outCh)
}
