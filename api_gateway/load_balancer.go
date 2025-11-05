package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
   For load balancing of services I will use Round Robin algorithm
   It is simple one but fit to my assumption about services here
   as all service instances are the same. in case of different instances
   capabilities weighted (dynamic) round robin can be used.
*/

type RoundRobin struct {
	ctx   context.Context
	idx   atomic.Uint32 // to ensure even distributed under concurrent use
	conns []*Service    // persistence conns
	// current atomic.Value  // atomic point to curr snapshot
}

type Service struct {
	conn    *grpc.ClientConn
	URL     string
	healthy atomic.Bool
}

/*
	    TODO:

		A good enchancment here would be add a watcher that watches config file
		and recreate the load balancer objects for services if any new urls added
		and to avoid concurrency issues, COPY ON WRITE can be used with just current atomic pointer for curr snapshot
		More complexity but cool and zero downtime for api_gateway
*/
func NewRoundRobin(ctx context.Context, serviceURLs []string, pingInterval time.Duration) *RoundRobin {
	if pingInterval <= 0 {
		pingInterval = 5 * time.Second
	}
	r := &RoundRobin{
		ctx:   ctx,
		conns: make([]*Service, 0, len(serviceURLs)),
	}

	for _, url := range serviceURLs {
		conn, err := grpc.NewClient(url,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Printf("error connecting to url: %s, err: %v", url, err)
			continue
		}
		// conn.Connect()
		srv := &Service{
			conn: conn,
			URL:  url,
		}
		r.conns = append(r.conns, srv)
		// log.Printf("Connected to service instance: %s", url)
	}

	if len(r.conns) == 0 {
		log.Fatal("No service instances available")
	}

	// intial health check at first
	r.healthCheck()

	go r.ping(pingInterval) // periodic health checks
	return r
}

func (r *RoundRobin) ServiceConn() (*grpc.ClientConn, error) {
	tries := 0
	maxTries := 2 * len(r.conns)

	for tries < maxTries {
		idx := r.idx.Add(1) % uint32(len(r.conns))
		tries++

		if r.conns[idx].healthy.Load() {
			return r.conns[idx].conn, nil
		}
	}

	return nil, errors.New("no healthy instance for this service now")
}

func (r *RoundRobin) ping(pingInterval time.Duration) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.healthCheck()
		}
	}
}

func (r *RoundRobin) healthCheck() {
	for _, srv := range r.conns {
		resp, err := http.Get(healthEnd(srv.URL))
		if err == nil && resp.StatusCode == http.StatusOK {
			if !srv.healthy.Load() {
				log.Printf("Service %s is now healthy", srv.URL)
			}
			// log.Printf("Service %s is healthy", srv.URL)
			srv.healthy.Store(true)
		} else {
			if srv.healthy.Load() {
				log.Printf("Service %s is now unhealthy: %v", srv.URL, err)
			}
			// log.Printf("Service %s is  unhealthy", srv.URL)
			srv.healthy.Store(false)
		}
	}
}

func (r *RoundRobin) Close() {
	for _, srv := range r.conns {
		if srv.conn != nil {
			srv.conn.Close()
		}
	}
}

func healthEnd(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return ""
	}

	hostname := parts[0]
	healthURL := fmt.Sprintf("http://%s:8080/health", hostname)

	return healthURL
}
