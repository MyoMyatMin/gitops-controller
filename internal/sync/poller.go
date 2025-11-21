package sync

import (
	"sync"
	"time"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/sirupsen/logrus"
)

type Poller struct {
	engine        *Engine
	interval      time.Duration
	stopCh        chan struct{}
	wg            *sync.WaitGroup
	lastCommitSHA string
}

func NewPoller(engine *Engine, interval time.Duration) *Poller {
	return &Poller{
		engine:   engine,
		interval: interval,
		stopCh:   make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}
}

func (p *Poller) Start() {

	log.Infof("Starting poller: checking for updates every %s", p.interval)

	retryConfig := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
	}

	p.wg.Add(1)
	defer p.wg.Done()
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info("Polling for changes....")
			latestSHA, err := p.engine.gitRepo.GetLatestCommit()
			if err != nil {
				log.Errorf("Error checking git commit: %v", err)
				continue
			}

			if p.lastCommitSHA != "" && p.lastCommitSHA == latestSHA {
				log.Info("No new commits found.")
				continue
			}

			log.WithFields(logrus.Fields{
				"new_commit": latestSHA,
				"old_commit": p.lastCommitSHA,
			}).Info("New commit found. Starting to sync.")

			result, err := p.engine.SyncWithRetry(retryConfig)
			if err != nil {
				log.Errorf("Sync failed: %v", err)
			} else {

				log.WithFields(logrus.Fields{
					"commit":  result.CommitSHA,
					"updated": len(result.Updated),
					"deleted": len(result.Deleted),
					"errors":  len(result.Errors),
				}).Info("Sync complete")
			}
			if err == nil {
				p.lastCommitSHA = latestSHA
			}

		case <-p.stopCh:
			log.Info("Stopping poller.")
			return
		}
	}
}

func (p *Poller) Stop() {
	log.Info("Sending stop signal to poller...")
	close(p.stopCh)
	p.wg.Wait()
	log.Info("Poller stopped successfully.")
}
