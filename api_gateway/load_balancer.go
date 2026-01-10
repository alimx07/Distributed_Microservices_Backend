package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/api_gateway/models"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
   For load balancing of services I will use Round Robin algorithm
   It is simple one but fit to my assumption about services here
   as all service instances are the same. in case of different instances
   capabilities weighted (dynamic) round robin can be used.
*/

type LoadBalancer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	Balancers map[string]*RoundRobin
	client    *etcd.Client
}

type RoundRobin struct {
	mu    sync.RWMutex
	idx   atomic.Uint32 // to ensure even distributed under concurrent use
	conns []*Service    // persistence conns
}

type Service struct {
	conn *grpc.ClientConn
	ID   string
	// healthy atomic.Bool
}

/*
	    TODO:

		A good enchancment here would be add a watcher that watches config file
		and recreate the load balancer objects for services if any new urls added
		More complexity but cool and zero downtime for api_gateway
*/

func NewLoadBalancer(config models.RegisteryConfig) (*LoadBalancer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	lb := &LoadBalancer{ctx: ctx, cancel: cancel, Balancers: make(map[string]*RoundRobin)}
	client, err := etcd.New(etcd.Config{
		Endpoints:   strings.Split(config.ServiceRegisteryPath, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("Error in intiallizing etcd Client: %v", err.Error())
	}
	lb.client = client
	lb.wg.Add(1)
	go func() {
		lb.watch(config.ServiceRegisteryPrefix)
		lb.wg.Done()
	}()
	return lb, nil
}

func (lb *LoadBalancer) watch(prefix string) {
	watchChan := lb.client.Watch(lb.ctx, prefix, etcd.WithPrefix())
	resp, err := lb.client.Get(lb.ctx, prefix, etcd.WithPrefix())
	if err != nil {
		log.Printf("Error in getting previous Keys in registery if any: %v", err)
	}
	if resp != nil {
		for _, entry := range resp.Kvs {
			key, val, ID := getKeyValID(entry)
			log.Printf("PUT %s = %s", key, val)
			if _, ok := lb.Balancers[key]; !ok {
				lb.Balancers[key] = newRoundRobin()
			}
			lb.Balancers[key].update(val, ID)
		}
	}
	for watchRes := range watchChan {
		if watchRes.Canceled {
			log.Printf("Etcd Watcher Failed: %v", watchRes.Err())
			break
		}
		if err := watchRes.Err(); err != nil {
			log.Printf("watch error: %v", err)
			continue
		}
		for _, ev := range watchRes.Events {
			key, val, ID := getKeyValID(ev.Kv)
			switch ev.Type {
			case etcd.EventTypePut:
				log.Printf("PUT %s = %s", key, val)
				if _, ok := lb.Balancers[key]; !ok {
					lb.Balancers[key] = newRoundRobin()
				}
				lb.Balancers[key].update(val, ID)
			case etcd.EventTypeDelete:
				log.Printf("DELETE instance in Service %s", key)
				lb.Balancers[key].delete(ID)
			}
		}
	}
}

func (lb *LoadBalancer) close() {
	lb.cancel()
	lb.client.Close()
	lb.wg.Wait()
	for _, lb := range lb.Balancers {
		lb.close()
	}
}

func newRoundRobin() *RoundRobin {
	// if pingInterval <= 0 {
	// 	pingInterval = 5 * time.Second
	// }
	r := &RoundRobin{
		conns: make([]*Service, 0),
	}

	// go r.ping(pingInterval) // periodic health checks
	return r
}

func (r *RoundRobin) update(serviceURL string, ID string) {
	conn, err := grpc.NewClient(serviceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("error connecting to url: %s, err: %v", serviceURL, err)
		return
	}

	srv := &Service{
		conn: conn,
		ID:   ID,
	}
	// lock
	r.mu.Lock()
	defer r.mu.Unlock()

	r.conns = append(r.conns, srv)
	log.Printf("New Instance Added: %v %v", serviceURL, ID)

}

func (r *RoundRobin) delete(ID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, srv := range r.conns {
		// log.Println(srv.ID, ID, srv.ID == ID)
		if srv.ID == ID {
			// close connection
			r.conns[i].conn.Close()
			r.conns[i] = r.conns[len(r.conns)-1]
			r.conns = r.conns[:len(r.conns)-1]
			return
		}
	}
	log.Printf("%v not found", ID)
}

func (r *RoundRobin) ServiceConn() (*grpc.ClientConn, error) {

	if len(r.conns) == 0 {
		return nil, errors.New("no healthy instance for this service now")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := r.idx.Add(1) % uint32(len(r.conns))
	return r.conns[idx].conn, nil

}

// func (r *RoundRobin) ping(pingInterval time.Duration) {

// 	ticker := time.NewTicker(pingInterval)
// 	defer ticker.Stop()
// 	client := &http.Client{Timeout: 2 * time.Second}
// 	// intial healthCheck
// 	r.healthCheck(client)

// 	//start checking prediocally
// 	for {
// 		select {
// 		case <-r.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			r.healthCheck(client)
// 		}
// 	}
// }

// func (r *RoundRobin) healthCheck(client *http.Client) {
// 	for _, srv := range r.conns {
// 		resp, err := client.Get(healthEnd(srv.URL))
// 		if err == nil && resp.StatusCode == http.StatusOK {
// 			if !srv.healthy.Load() {
// 				log.Printf("Service %s is now healthy", srv.URL)
// 			}
// 			// log.Printf("Service %s is healthy", srv.URL)
// 			srv.healthy.Store(true)
// 			resp.Body.Close() // close body
// 		} else {
// 			if srv.healthy.Load() {
// 				log.Printf("Service %s is now unhealthy: %v", srv.URL, err)
// 			}
// 			// log.Printf("Service %s is  unhealthy", srv.URL)
// 			srv.healthy.Store(false)
// 		}
// 	}

// }

func (r *RoundRobin) close() {

	// End Ping
	// r.cancel()
	for _, srv := range r.conns {
		if srv.conn != nil {
			srv.conn.Close()
		}
	}
}

func getKeyValID(kv *mvccpb.KeyValue) (string, string, string) {
	key, val := strings.Split(string(kv.Key), "/"), string(kv.Value)
	return key[2], val, key[3]
}

// func getKeyID(ev *etcd.Event) (string, string) {
// 	key := strings.Split(string(ev.Kv.Key), "/")
// 	return key[2], key[3]
// }

// func healthEnd(addr string) string {
// 	parts := strings.Split(addr, ":")
// 	if len(parts) != 2 {
// 		return ""
// 	}

// 	hostname := parts[0]
// 	healthURL := fmt.Sprintf("http://%s:8080/health", hostname)

// 	return healthURL
// }
