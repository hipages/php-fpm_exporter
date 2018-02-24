// Copyright © 2018 Enrico Stahn <enrico.stahn@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/hipages/php-fpm_exporter/phpfpm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

// Configuration variables
var (
	listeningAddress  string
	metricsEndpoint   string
	scrapeURIs        []string
	customLabelNames  []string
	customLabelValues []string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Starting server on %v with path %v", listeningAddress, metricsEndpoint)

		pm := phpfpm.PoolManager{}

		for _, uri := range scrapeURIs {
			pm.Add(uri)
		}

		exporter := phpfpm.NewExporter(pm)
		prometheus.MustRegister(exporter)

		http.Handle(metricsEndpoint, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			 <head><title>Apache Exporter</title></head>
			 <body>
			 <h1>Apache Exporter</h1>
			 <p><a href='` + metricsEndpoint + `'>Metrics</a></p>
			 </body>
			 </html>`))
		})
		log.Fatal(http.ListenAndServe(listeningAddress, nil))
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serverCmd.Flags().StringVar(&listeningAddress, "web.listen-address", ":9253", "Address on which to expose metrics and web interface.")
	serverCmd.Flags().StringVar(&metricsEndpoint, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	serverCmd.Flags().StringSliceVar(&scrapeURIs, "phpfpm.scrape-uri", []string{"tcp://127.0.0.1:9000/status"}, "FastCGI address, e.g. unix:///tmp/php.sock;/status or tcp://127.0.0.1:9000/status")

	//viper.BindEnv("web.listen-address", "PHP_FPM_WEB_LISTEN_ADDRESS")
	//viper.BindPFlag("web.listen-address", serverCmd.Flags().Lookup("web.listen-address"))

	// Workaround since vipers BindEnv is currently not working as expected (see https://github.com/spf13/viper/issues/461)

	envs := map[string]string{
		"PHP_FPM_WEB_LISTEN_ADDRESS": "web.listen-address",
		"PHP_FPM_WEB_TELEMETRY_PATH": "web.telemetry-path",
		"PHP_FPM_SCRAPE_URI":         "phpfpm.scrape-uri",
	}

	for env, flag := range envs {
		flag := serverCmd.Flags().Lookup(flag)
		flag.Usage = fmt.Sprintf("%v [env %v]", flag.Usage, env)
		if value := os.Getenv(env); value != "" {
			flag.Value.Set(value)
		}
	}
}
