package agscheduler

import (
	"fmt"
	"log/slog"
	"net/rpc"
	"strings"

	"github.com/google/uuid"
)

type Main struct {
	Id                string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
}

type Worker struct {
	Id                string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
}

type ClusterMain struct {
	Id                string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
	queueMap          map[string]map[string]map[string]any
}

func (cm *ClusterMain) SetId() {
	cm.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

func (cm *ClusterMain) Register(w *Worker, m *Main) error {
	slog.Info(fmt.Sprintf("Registration from the cluster worker `%s:%s`:", w.Id, w.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Worker Scheduler RPC Service listening at: %s", w.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Worker Scheduler RPC Service queue: `%s`", w.SchedulerQueue))

	if cm.queueMap == nil {
		cm.queueMap = make(map[string]map[string]map[string]any)
	}
	if _, ok := cm.queueMap[cm.SchedulerQueue]; !ok {
		cm.queueMap[cm.SchedulerQueue] = map[string]map[string]any{}
	}
	cm.queueMap[cm.SchedulerQueue][cm.Id] = map[string]any{
		"endpoint":           cm.Endpoint,
		"scheduler_endpoint": cm.SchedulerEndpoint,
		"health":             true,
	}
	if _, ok := cm.queueMap[w.SchedulerQueue]; !ok {
		cm.queueMap[w.SchedulerQueue] = map[string]map[string]any{}
	}
	cm.queueMap[w.SchedulerQueue][w.Id] = map[string]any{
		"endpoint":           w.Endpoint,
		"scheduler_endpoint": w.SchedulerEndpoint,
		"health":             true,
	}

	m.Id = cm.Id
	m.Endpoint = cm.Endpoint
	m.SchedulerEndpoint = cm.SchedulerEndpoint
	m.SchedulerQueue = cm.SchedulerQueue

	return nil
}

type ClusterWorker struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
}

func (cw *ClusterWorker) SetId() {
	cw.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

func (cw *ClusterWorker) Register() error {
	slog.Info(fmt.Sprintf("Register with cluster main `%s`:", cw.MainEndpoint))

	rClient, err := rpc.DialHTTP("tcp", cw.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main: `%s`, error: %s", cw.MainEndpoint, err)
	}

	worker := Worker{Id: cw.Id, Endpoint: cw.Endpoint, SchedulerEndpoint: cw.SchedulerEndpoint, SchedulerQueue: cw.SchedulerQueue}
	var main Main
	err = rClient.Call("CRPCService.Register", worker, &main)
	if err != nil {
		return fmt.Errorf("failed to register to cluster main, error: %s", err)
	}

	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service queue: `%s`", main.SchedulerQueue))

	return nil
}
