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

package register

import (
	"testing"
	"fmt"
	"strings"
	etcd3 "github.com/coreos/etcd/clientv3"
	"context"
	"goservice"
)

func Test_Register(t *testing.T) {
	fmt.Println("asdasda")
	endpoints := []string{"119.23.67.84:2379", "120.79.38.97:2379"}
	Register("TestService", "192.168.0.1", 8080, endpoints, 10, 10)
	//Register()
}

func Test_GetData(t *testing.T) {
	client, err := etcd3.New(etcd3.Config{
		Endpoints: strings.Split("119.23.67.84:2379,120.79.38.97:2379", ","),
	})
	if err != nil {
		fmt.Println("Create etcdClient failed: ", err.Error())
		return
	}
	key := fmt.Sprintf("/%s/%s", service.Prefix, "TestServic2")
	// should get first, if not exist, set it
	resp, err := client.Get(context.Background(), key, etcd3.WithPrefix())
	if err != nil {
		fmt.Println("Get Data failed: ", err.Error())
		return
	}
	serviceList := service.ExtractAddress(resp)
	fmt.Println(serviceList)
}
