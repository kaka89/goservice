/* Copyright 2018 Bruce Liu.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	etcd3 "github.com/coreos/etcd/clientv3"
	"fmt"
	"errors"
	"math/rand"
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"log"
	"goservice"
)

// 基本方式：本地维护一个服务发现的结果，优先从本地服务读取值. 第一次获取一个服务的地址时，因本地没有，故从etcd服务获取，然后换到到本地，并同时启动该服务的watcher，有更新时，及时更新本地缓存。
// Tips 一个服务需要发现多个其他服务，故需要有client，且后期通过client来获取服务的值
type DiscoveryClient struct {
	etcdClient    *etcd3.Client
	EtcdEndpoints []string

	// 服务列表: key 为service 名字
	serviceMap map[string][]string

	watcher map[string]bool
}

func NewClient(etcdEndpoints []string) (*DiscoveryClient, error) {
	client := &DiscoveryClient{}
	client.EtcdEndpoints = etcdEndpoints
	etcdClient, err := etcd3.New(etcd3.Config{
		Endpoints: client.EtcdEndpoints,
	})
	if err != nil {
		return nil, fmt.Errorf("Create register etcdClient failed: %s", err.Error())
	}
	client.etcdClient = etcdClient

	client.serviceMap = make(map[string][]string)

	client.watcher = make(map[string]bool)

	return client, nil
}

func (c *DiscoveryClient) GetService(serviceName string) (string, error) {

	if serviceName == "" {
		return "", errors.New("Service name should not be null")
	}

	// 从本地缓存的获取
	if len(c.serviceMap[serviceName]) > 0 {
		i := rand.Intn(len(c.serviceMap[serviceName]))
		return c.serviceMap[serviceName][i], nil
	}

	// 从 etcd服务获取
	resp, err := c.etcdClient.Get(context.Background(), c.getKey(serviceName), etcd3.WithPrefix())
	if err != nil {
		log.Printf("Get service failed: %s, error: %s", serviceName, err.Error())
		return "", err
	}

	// 建立watcher，有变化，及时更新 只启动一个watch?
	if !c.watcher[serviceName] {
		go c.watch(serviceName)
	}

	serviceList := service.ExtractAddress(resp)

	if len(serviceList) == 0 {
		return "", errors.New("Service not found.")
	}

	c.serviceMap[serviceName] = serviceList

	i := rand.Intn(len(c.serviceMap[serviceName]))
	return c.serviceMap[serviceName][i], nil
}

func (c *DiscoveryClient) getKey(serviceName string) string {
	return fmt.Sprintf("/%s/%s", service.Prefix, serviceName)
}

// 开始watch这个service的注册信息
func (c *DiscoveryClient) watch(serviceName string) {
	c.watcher[serviceName] = true
	defer func() {
		// 如果此函数退出，那么则可以重新建立watcher
		c.watcher[serviceName] = false
	}()
	// prefix is the etcd prefix/value to watch
	prefix := c.getKey(serviceName)
	log.Printf("Watch service : %s", serviceName)
	// 创建watch
	rch := c.etcdClient.Watch(context.Background(), prefix, etcd3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.Printf("Catch changes: %s, event: %d", serviceName, ev.Type)
			switch ev.Type {
			case mvccpb.PUT:
				find := false
				for _, v := range c.serviceMap[serviceName] {
					if v == string(ev.Kv.Value) {
						find = true
						break
					}
				}
				if !find {
					log.Printf("Add new service: %s, value: %s", serviceName, string(ev.Kv.Value))
					c.serviceMap[serviceName] = append(c.serviceMap[serviceName], string(ev.Kv.Value))
				}
				break
			case mvccpb.DELETE:
				index := -1
				fmt.Println(string(ev.Kv.Value))
				key := string(ev.Kv.Key)
				value := key[len(c.getKey(serviceName))+1:len(key)]
				fmt.Println(value)

				for i, v := range c.serviceMap[serviceName] {
					if v == value {
						index = i
						break
					}
				}

				if index >= 0 && index < len(c.serviceMap[serviceName]) {
					log.Printf("Delete service: %s, value: %s", serviceName, string(ev.Kv.Value))
					c.serviceMap[serviceName] = append(c.serviceMap[serviceName][0:index], c.serviceMap[serviceName][index+1: len(c.serviceMap[serviceName])]...)
				}
				break
			}
		}
	}

}
