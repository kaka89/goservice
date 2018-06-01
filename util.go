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

package service

import (
	etcd3 "github.com/coreos/etcd/clientv3"
)

var (
	// 默认的prefix
	Prefix = "service"
)

func ExtractAddress(resp *etcd3.GetResponse) []string {
	addresses := []string{}

	if resp == nil || resp.Kvs == nil {
		return addresses
	}

	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			addresses = append(addresses, string(v))
		}
	}

	return addresses
}
