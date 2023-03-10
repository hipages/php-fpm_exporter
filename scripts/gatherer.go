// Copyright 2017 Kumina, https://kumina.nl/
// Copyright 2023 Guilhem Lettron, guilhem@barpilot.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package scripts

package scripts

import (
	"fmt"
	"net/url"
	"path"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	fcgiclient "github.com/tomasen/fcgi_client"
)

var (
	phpfpmSocketPathLabel = "socket_path"
	phpfpmScriptPathLabel = "script_path"
)

// Gathered collects metrics from scripts via PHP-FPM.
type Gathered struct {
	Sockets []string
	Scripts []string
}

// Gather collects metrics from scripts via PHP-FPM.
func (g *Gathered) Gather() ([]*dto.MetricFamily, error) {
	var result []*dto.MetricFamily

	for _, socketPath := range g.Sockets {
		parsedURL, err := url.Parse(socketPath)
		if err != nil {
			return result, fmt.Errorf("failed to parse socket path %q: %s", socketPath, err)
		}

		for _, scriptPath := range g.Scripts {
			fcgi, err := fcgiclient.Dial(parsedURL.Scheme, path.Join(parsedURL.Host, parsedURL.Path))
			if err != nil {
				return result, fmt.Errorf("failed to connect to %q PHP-FPM socket %q: %w", parsedURL.Scheme, parsedURL.Host, err)
			}
			defer fcgi.Close()

			env := make(map[string]string)
			env["DOCUMENT_ROOT"] = path.Dir(scriptPath)
			env["SCRIPT_FILENAME"] = scriptPath
			env["SCRIPT_NAME"] = path.Base(scriptPath)
			env["REQUEST_METHOD"] = "GET"

			resp, err := fcgi.Get(env)
			if err != nil {
				return result, fmt.Errorf("failed to get metrics from PHP-FPM socket %q: %w", socketPath, err)
			}

			var parser expfmt.TextParser
			parsed, err := parser.TextToMetricFamilies(resp.Body)
			if err != nil {
				return result, fmt.Errorf("failed to parse metrics from PHP-FPM socket %q: %w", socketPath, err)
			}

			for _, metricFamily := range parsed {
				for _, metric := range metricFamily.Metric {
					socketPathCopy := parsedURL.String()
					socketLabel := &dto.LabelPair{
						Name:  &phpfpmSocketPathLabel,
						Value: &socketPathCopy,
					}

					scriptPathCopy := scriptPath
					scriptLabel := &dto.LabelPair{
						Name:  &phpfpmScriptPathLabel,
						Value: &scriptPathCopy,
					}

					metric.Label = append(metric.Label, socketLabel, scriptLabel)
				}
				result = append(result, metricFamily)
			}
		}
	}

	return result, nil
}
