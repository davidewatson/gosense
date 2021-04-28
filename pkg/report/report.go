// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package report

// TODO(dwat): The report(s) could use some work. For example it would be
// helpful if future formats reported the version of gosense used to generate
// the data. We might also want to consider supporting the Prometheus v2 format
// for easier integration with ODS and Kubernetes:
// https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md
// https://github.com/prometheus/client_golang

// ClassicReport is the original monitoring API.
type ClassicReport struct {
	Information []map[string]string `json:"Information"`
	Actions     []map[string]string `json:"Actions"`
	Resources   []map[string]string `json:"Resources"`
}
