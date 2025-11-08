package sync

import (
	"fmt"
	"sync"
	"time"
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
	fmt.Printf("Starting poller: checking for updates every %s\n", p.interval)

	p.wg.Add(1)
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Println("Polling for changes....")

			latestSHA, err := p.engine.gitRepo.GetLatestCommit()
			if err != nil {
				fmt.Printf("Error getting latest commit: %v\n", err)
				continue
			}

			if p.lastCommitSHA != "" && p.lastCommitSHA == latestSHA {
				fmt.Println("No new commits found.")
				continue
			}

			fmt.Printf("New commit %s found (was %s). Starting to sync.\n", latestSHA, p.lastCommitSHA)
			result, err := p.engine.Sync()
			if err != nil {
				fmt.Printf("Sync failed: %v\n", err)
			} else {
				printSyncResult(*result)
			}

			p.lastCommitSHA = latestSHA

		case <-p.stopCh:
			fmt.Println("Stopping poller.")
			ticker.Stop()
			return
		}
	}
}

func (p *Poller) Stop() {
	fmt.Println("Sending stop signal to poller...")
	close(p.stopCh)
	p.wg.Wait()
	fmt.Println("Poller stopped.")
}

func printSyncResult(r SyncResult) {
	fmt.Printf("Sync to commit %s complete.\n", r.CommitSHA)
	fmt.Printf("- Updated: %d\n", len(r.Updated))
	fmt.Printf("- Deleted: %d\n", len(r.Deleted))
	if len(r.Errors) > 0 {
		fmt.Printf("- Errors: %d\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}
}
