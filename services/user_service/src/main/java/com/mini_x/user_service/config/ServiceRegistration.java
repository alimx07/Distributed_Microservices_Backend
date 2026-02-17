package com.mini_x.user_service.config;

import java.nio.charset.StandardCharsets;
import java.util.UUID;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import io.etcd.jetcd.ByteSequence;
import io.etcd.jetcd.Client;
import io.etcd.jetcd.options.PutOption;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;

@Component
public class ServiceRegistration {

    private static final Logger logger = LoggerFactory.getLogger(ServiceRegistration.class);
    private static final String SERVICE_PREFIX = "/services/user_service/";

    private final Client etcdClient;
    private final String grpcPort;
    private final String hostname;
    private final long leaseTtl;
    private final String instanceId;
    
    private long leaseId;

    public ServiceRegistration(
            Client etcdClient,
            @Value("${grpc.server.port}") String grpcPort,
            @Value("${HOSTNAME:unknown}") String hostname,
            @Value("${etcd.lease-ttl:5}") long leaseTtl) {
        this.etcdClient = etcdClient;
        this.grpcPort = grpcPort;
        this.hostname = hostname;
        this.leaseTtl = leaseTtl;
        this.instanceId = UUID.randomUUID().toString();
    }

    @PostConstruct
    public void register() {
        try {
            String serviceAddress = hostname + ":" + grpcPort;
            String serviceKey = SERVICE_PREFIX + instanceId;

            leaseId = etcdClient.getLeaseClient().grant(leaseTtl).get().getID();

            ByteSequence key = ByteSequence.from(serviceKey, StandardCharsets.UTF_8);
            ByteSequence value = ByteSequence.from(serviceAddress, StandardCharsets.UTF_8);
            PutOption putOption = PutOption.builder().withLeaseId(leaseId).build();

            etcdClient.getKVClient().put(key, value, putOption).get();
            etcdClient.getLeaseClient().keepAlive(leaseId, new io.grpc.stub.StreamObserver<>() {
                @Override public void onNext(io.etcd.jetcd.lease.LeaseKeepAliveResponse r) {}
                @Override public void onError(Throwable t) {}
                @Override public void onCompleted() {}
            });

            logger.info("Service registered: key={}, address={}", serviceKey, serviceAddress);
        } catch (Exception e) {
            logger.error("Failed to register service: {}", e.getMessage());
        }
    }

    @PreDestroy
    public void unregister() {
        try {
            if (leaseId != 0) {
                etcdClient.getLeaseClient().revoke(leaseId).get();
            }
            etcdClient.close();
        } catch (Exception e) {
            logger.error("Failed to unregister service: {}", e.getMessage());
        }
    }
}
