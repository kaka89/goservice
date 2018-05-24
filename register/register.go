package register

import (
	"fmt"
	"time"
	etcd3 "github.com/coreos/etcd/clientv3"
	"context"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"log"
	"goservice"
)

type RegisterClient struct {
	etcdClient      *etcd3.Client
	EtcdEndpoints   []string
	refreshInterval time.Duration
	keyTTL          int
	serviceName     string
	host            string
	port            int
	stopSignal      chan bool
}

// 服务注册，命名方式"/service/{serviceName}/{serviceHost}:{servicePort}", 值为："{serviceHost}:{servicePort}"， 注册是当前服务本身而已，故可直接注册
//Tips: 如果需要，可以将值的内容转换成json格式，增减内容丰富度
func Register(serviceName string, host string, port int, EtcdEndpoints []string, refreshInterval time.Duration, keyTTL int) (*RegisterClient, error) {
	client := &RegisterClient{}
	client.serviceName = serviceName
	client.host = host
	client.port = port
	client.EtcdEndpoints = EtcdEndpoints
	client.refreshInterval = refreshInterval
	client.keyTTL = keyTTL
	// 保证存活时间比刷新时间长
	if keyTTL < int(refreshInterval.Seconds()) {
		keyTTL = 2 * int(refreshInterval.Seconds())
	}

	etcdClient, err := etcd3.New(etcd3.Config{
		Endpoints: client.EtcdEndpoints,
	})
	if err != nil {
		return nil, fmt.Errorf("Create etcd client failed: s%s", err)
	}
	client.stopSignal = make(chan bool, 1)
	client.etcdClient = etcdClient
	client.register()
	return client, nil
}

func (c *RegisterClient) register() error {

	go func() {
		// invoke self-register with ticker
		ticker := time.NewTicker(c.refreshInterval)
		for {
			// should get first, if not exist, set it
			resp, err := c.etcdClient.Grant(context.Background(), int64(c.keyTTL))
			if err != nil {
				log.Printf("Register service failed. serviceName: %s, err: %s", c.serviceName, err.Error())
			}
			fmt.Println(c.getKey())
			_, err = c.etcdClient.Get(context.Background(), c.getKey())
			value := fmt.Sprintf("%s:%d", c.host, c.port)

			if err != nil {
				if err == rpctypes.ErrKeyNotFound {
					if _, err := c.etcdClient.Put(context.Background(), c.getKey(), value, etcd3.WithLease(resp.ID)); err != nil {
						//if _, err := c.etcdClient.Put(context.Background(), c.getKey(), value); err != nil {
						log.Printf("Register service failed. serviceName: %s, value: %s, err: %s", c.serviceName, value, err.Error())
					}
				} else {
					log.Printf("Get Value from etcd failed: %s", err)
				}
			} else {

				if _, err := c.etcdClient.Put(context.Background(), c.getKey(), value, etcd3.WithLease(resp.ID)); err != nil {
					//if _, err := c.etcdClient.Put(context.Background(), c.getKey(), value); err != nil {
					log.Printf("Register service failed. serviceName: %s, value: %s, err: %s", c.serviceName, value, err.Error())
				} else {
					log.Printf("Register current service: %s, value: %s", c.serviceName, value)
				}
			}
			select {
			case <-c.stopSignal: // 手动停止注册
				return
			case <-ticker.C: // 等待一定时间间隔
			}
		}
	}()
	return nil
}

// 只有服务在注册之后，才能够取消注册
func (c *RegisterClient) UnRegister() error {
	c.stopSignal <- true

	_, err := c.etcdClient.Get(context.Background(), c.getKey())

	if err != nil {
		log.Printf("Get Value from etcd failed: %s", err.Error())
	} else if _, err := c.etcdClient.Delete(context.Background(), c.getKey()); err != nil {
		log.Printf("UnRegister service failed. serviceName: %s, err: %s", c.serviceName, err.Error())
	}
	return err
}

func (c *RegisterClient) getKey() string {
	return fmt.Sprintf("/%s/%s/%s:%d", service.Prefix, c.serviceName, c.host, c.port)
}
