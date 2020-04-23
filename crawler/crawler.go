package crawler

import (
	"context"
	"github.com/angrymuskrat/instagram-auditor/crawler/data"
	"github.com/visheratin/unilog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

const PackSize = 1000

type entity struct {
	id      string
	err     error
	profile *data.Profile
}

type Crawler struct {
	workers         []*worker
	db              *mongo.Client
	profilesCollect *mongo.Collection
	inCh            chan entity
	outCh           chan entity
}

func New(ctx context.Context, configPath string) *Crawler {
	cfg, err := readConfig(configPath)
	if err!= nil {
		return nil
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongoUrl))
	if err != nil {
		unilog.Logger().Error("don't be able to create mongo client", zap.Error(err))
		return nil
	}

	err = client.Connect(ctx)
	if err != nil {
		unilog.Logger().Error("don't be able to connect with mongo", zap.Error(err))
		return nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		unilog.Logger().Error("don't be able to connect with mongo", zap.Error(err))
	}
	collection := client.Database("instagramAuditor").Collection("profiles")

	cr := Crawler{
		inCh:  make(chan entity),
		outCh: make(chan entity),
		db:    client,
		profilesCollect: collection,
	}
	cr.workers = make([]*worker, len(cfg.TorPorts))
	for i, p := range cfg.TorPorts {
		cr.workers[i] = &worker{
			id:    i,
			inCh:  cr.inCh,
			outCh: cr.outCh,
		}
		cr.workers[i].init(p)
		go cr.workers[i].start()
	}
	return &cr
}

func (c *Crawler) Start(ctx context.Context, ids []string) (brokenIds []string) {
	go c.distributeEntities(ids)

	defer func () {
		err := c.db.Disconnect(ctx)
		if err != nil {
			unilog.Logger().Error("don't be able to disconnect with mongo", zap.Error(err))
		}
	}()

	var profiles []interface{}
	proceed := 0
	collect := 0
	for e := range c.outCh {
		proceed++
		if e.err != nil {
			brokenIds = append(brokenIds, e.id)
			continue
		}
		collect++
		profiles = append(profiles, *e.profile)
		if len(profiles) >= PackSize {
			err := c.saveProfiles(ctx, profiles, collect, len(brokenIds))
			if err != nil {
				return
			}
			profiles = []interface{}{}
		}
	}
	err := c.saveProfiles(ctx, profiles, collect, len(brokenIds))
	if err != nil {
		return
	}
	return
}

func (c *Crawler) saveProfiles(ctx context.Context, profiles []interface{},  collect, broken int) error {
	_, err := c.profilesCollect.InsertMany(ctx, profiles)
	if err != nil {
		unilog.Logger().Error("don't be able to save profiles", zap.Error(err))
		return err
	}
	profiles = []interface{}{}

	unilog.Logger().Info("status",
		zap.Int("collected", collect),
		zap.Int("broken ids", broken),
	)
	return nil
}

func (c *Crawler) distributeEntities(ids []string) {
	for _, id := range ids {
		c.inCh <- entity{id: id}
	}
	time.Sleep(1 * time.Second)
	close(c.inCh)
	close(c.outCh)
}
