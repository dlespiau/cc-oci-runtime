// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"json"
	"github.com/01org/tests/reporter/metric"
)

// Expected on-disk layout
//
//  fedora24-clear/environment.json	<- Description of the environment.
//  fedora24-clear/log/build.txt	<- Log of the operation build.json is from.
//					   Can be a symlink if several test
//					   results are derived from the same log.
//  fedora24-clear/build.json		<- Metric (boolean)
//  fedora24-clear/make-check.json	<- Metric (bookean)
//  fedora24-clear/docker.json		<- Metric (test-triplet, ie. passed/failed/skipped)
//  fedora24-clear/benchmark.json	<- Metric (number)
//  ubuntu16.04-clear/enviroment.json
//  .
//  .
//  .
//  report.json				<- description of the final report

// eg. "Fedora 24/Clear kernel & VM"
type enviroment {
	Name string `json:"name"`

}

type report struct {
	Environments []enviroment `json:"enviroments"`
	MetricDescs []metric.Desc `json:"metrics"`
	Metrics []metric.Metric
}

func main() {

}
