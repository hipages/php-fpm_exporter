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
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hipages/php-fpm_exporter/phpfpm"
	"github.com/hipages/php-fpm_exporter/scripts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// Configuration variables
var (
	listeningAddress string
	metricsEndpoint  string
	scrapeURIs       []string
	fixProcessCount  bool
	scriptsPaths     []string
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

		if fixProcessCount {
			log.Info("Idle/Active/Total Processes will be calculated by php-fpm_exporter.")
			exporter.CountProcessState = true
		}

		prometheus.MustRegister(exporter)

		srv := &http.Server{
			Addr: listeningAddress,
			// Good practice to set timeouts to avoid Slowloris attacks.
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		}

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
		}

		// add scripts gatherer if scripts are defined
		if len(scriptsPaths) > 0 {

			addresses := make([]string, len(scrapeURIs))
			for i, uri := range scrapeURIs {
				addresses[i] = strings.SplitN(uri, ";", 2)[0]
			}

			log.Debugf("Scripts to execute: %v", scriptsPaths)
			log.Debugf("Scrape URIs: %v", addresses)

			sGather := &scripts.Gathered{
				Scripts: scriptsPaths,
				Sockets: addresses,
			}

			gatherers = append(gatherers, sGather)
		}

		http.Handle(metricsEndpoint, promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{}))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`<html>
			 <head><title>php-fpm_exporter</title></head>
			 <body>
			 <h1>php-fpm_exporter</h1>
			 <p><a href='` + metricsEndpoint + `'>Metrics</a></p>
			 </body>
			 </html>`))

			if err != nil {
				log.Error()
			}
		})

		// Run our server in a goroutine so that it doesn't block.
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Error(err)
			}
		}()

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM
		// SIGKILL, SIGQUIT will not be caught.
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Block until we receive our signal.
		<-c

		// Create a deadline to wait for.
		wait := time.Second * 10
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline.
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Error during shutdown", err)
		}
		// Optionally, you could run srv.Shutdown in a goroutine and block on
		// <-ctx.Done() if your application should wait for other services
		// to finalize based on context cancellation.
		log.Info("Shutting down")
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVar(&listeningAddress, "web.listen-address", ":9253", "Address on which to expose metrics and web interface.")
	serverCmd.Flags().StringVar(&metricsEndpoint, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	serverCmd.Flags().StringSliceVar(&scrapeURIs, "phpfpm.scrape-uri", []string{"tcp://127.0.0.1:9000/status"}, "FastCGI address, e.g. unix:///tmp/php.sock;/status or tcp://127.0.0.1:9000/status")
	serverCmd.Flags().BoolVar(&fixProcessCount, "phpfpm.fix-process-count", false, "Enable to calculate process numbers via php-fpm_exporter since PHP-FPM sporadically reports wrong active/idle/total process numbers.")
	serverCmd.Flags().StringSliceVar(&scriptsPaths, "scripts.path", []string{}, "PHP scripts to execute to get additional metrics.")

	// Workaround since vipers BindEnv is currently not working as expected (see https://github.com/spf13/viper/issues/461)

	envs := map[string]string{
		"PHP_FPM_WEB_LISTEN_ADDRESS": "web.listen-address",
		"PHP_FPM_WEB_TELEMETRY_PATH": "web.telemetry-path",
		"PHP_FPM_SCRAPE_URI":         "phpfpm.scrape-uri",
		"PHP_FPM_FIX_PROCESS_COUNT":  "phpfpm.fix-process-count",
	}

	mapEnvVars(envs, serverCmd)
}
