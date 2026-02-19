package main

//###################################
// NOTE: All code in comment was used during local deploy game
// Now, K8s Services handle load balancing via kube-proxy/iptables.
// Gateway now holds gRPC connections to K8s service DNS names aka serviceAccounts.
// Path: Ingress -> Gateway -> K8s Service (DNS) -> Pod
//###################################

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceConnections struct {
	conns map[string]*grpc.ClientConn
}

func NewServiceConnections(k8sServices map[string]string) (*ServiceConnections, error) {
	sc := &ServiceConnections{conns: make(map[string]*grpc.ClientConn)}
	for serviceName, addr := range k8sServices {
		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Printf("failed to connect to %s at %s: %v", serviceName, addr, err)
			continue
		}
		sc.conns[serviceName] = conn
		log.Printf("Connected to K8s service: %s -> %s", serviceName, addr)
	}
	if len(sc.conns) == 0 {
		return nil, fmt.Errorf("no service connections established")
	}
	return sc, nil
}

func (sc *ServiceConnections) GetConn(serviceName string) (*grpc.ClientConn, error) {
	conn, exists := sc.conns[serviceName]
	if !exists {
		return nil, fmt.Errorf("no connection for service: %s", serviceName)
	}
	return conn, nil
}

func (sc *ServiceConnections) close() {
	for name, conn := range sc.conns {
		if err := conn.Close(); err != nil {
			log.Printf("error closing connection to %s: %v", name, err)
		}
	}
}

//==============================
// OLD ONE
//==============================

// type LoadBalancer struct {
// 	// ctx       context.Context
// 	// cancel    context.CancelFunc
// 	// wg        sync.WaitGroup
// 	Balancers map[string]*RoundRobin
// 	// client    *etcd.Client
// }

// type RoundRobin struct {
// 	mu    sync.RWMutex
// 	idx   atomic.Uint32
// 	conns []*Service
// }

// type Service struct {
// 	conn *grpc.ClientConn
// 	ID   string
// }

// func NewLoadBalancer(k8sServices map[string]string) (*LoadBalancer, error) {
// 	lb := &LoadBalancer{Balancers: make(map[string]*RoundRobin)}
// 	for serviceName, addr := range k8sServices {
// 		conn, err := grpc.NewClient(addr,
// 			grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		)
// 		if err != nil {
// 			log.Printf("failed to connect to %s at %s: %v", serviceName, addr, err)
// 			continue
// 		}
// 		rr := &RoundRobin{conns: []*Service{{conn: conn, ID: serviceName}}}
// 		lb.Balancers[serviceName] = rr
// 		log.Printf("Connected to service: %s -> %s", serviceName, addr)
// 	}
// 	return lb, nil
// }

// func (lb *LoadBalancer) close() {
// 	for _, b := range lb.Balancers {
// 		b.close()
// 	}
// }

// func (r *RoundRobin) close() {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	for _, srv := range r.conns {
// 		srv.conn.Close()
// 	}
// }

// func (r *RoundRobin) ServiceConn() (*grpc.ClientConn, error) {
// 	if len(r.conns) == 0 {
// 		return nil, errors.New("no healthy instance for this service now")
// 	}
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()
// 	idx := r.idx.Add(1) % uint32(len(r.conns))
// 	return r.conns[idx].conn, nil
// }

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

// func (r *RoundRobin) close() {
// 	// End Ping
// 	// r.cancel()
// 	for _, srv := range r.conns {
// 		if srv.conn != nil {
// 			srv.conn.Close()
// 		}
// 	}
// }

// func getKeyValID(kv *mvccpb.KeyValue) (string, string, string) {
// 	key, val := strings.Split(string(kv.Key), "/"), string(kv.Value)
// 	return key[2], val, key[3]
// }

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
