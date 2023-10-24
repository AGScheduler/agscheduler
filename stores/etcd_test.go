package stores

import (
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kwkwc/agscheduler"
)

func TestEtcdStore(t *testing.T) {
	cli, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	defer cli.Close()
	store := &EtcdStore{
		Cli:          cli,
		JobsPath:     "/agscheduler/test_jobs",
		RunTimesPath: "/agscheduler/test_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)

	store.Clear()
}
