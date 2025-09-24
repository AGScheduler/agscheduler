package stores

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdStore(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(t, err)
	defer cli.Close()

	store := &EtcdStore{
		Cli:          cli,
		JobsPath:     "/agscheduler/test_jobs",
		RunTimesPath: "/agscheduler/test_run_times",
	}

	runTest(t, store)
}
