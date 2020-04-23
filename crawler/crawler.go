package crawler

import (
	"github.com/visheratin/unilog"
	"go.uber.org/zap"
	"time"
)

type Crawler struct {
	worker *worker
}

func New(port int) *Crawler {
	c := Crawler{}
	w := worker{}
	w.init(port)
	c.worker = &w
	return &c
}

func (c *Crawler) Start(entities []string, numPostsPerProfile int) {
	for _, id := range entities {
		nick, err := c.worker.getNickname(id)
		time.Sleep(time.Millisecond * 50)
		if err != nil {
			unilog.Logger().Error("don't be able to get nickname", zap.Error(err))
			continue
		}
		p, err := c.worker.getProfile(nick, id, numPostsPerProfile)
		time.Sleep(time.Millisecond * 50)
		if err != nil {
			unilog.Logger().Error("don't be able to get profile", zap.Error(err))
		} else {
			unilog.Logger().Info("success collect", zap.String("nickname", p.Username))
		}
	}
}
