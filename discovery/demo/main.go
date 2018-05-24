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

package main

import (
	"time"
	"fmt"
	"goservice/discovery"
)

func main() {
	endpoints := []string{"119.23.67.84:2379", "120.79.38.97:2379"}
	client, _ := discovery.NewClient(endpoints)
	for {
		serviceHost, err := client.GetService("TestService2")
		time.Sleep(time.Second)
		if err != nil {
			fmt.Println("error......", err)
			continue
		}
		fmt.Println("service: TestService2, value: ", serviceHost)
	}

	time.Sleep(time.Second * 500)
}
