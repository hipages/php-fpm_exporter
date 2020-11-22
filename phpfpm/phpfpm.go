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

// Package phpfpm provides convenient access to PHP-FPM pool data
package phpfpm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	fcgiclient "github.com/tomasen/fcgi_client"
)

// PoolProcessRequestIdle defines a process that is idle.
const PoolProcessRequestIdle string = "Idle"

// PoolProcessRequestRunning defines a process that is running.
const PoolProcessRequestRunning string = "Running"

// PoolProcessRequestFinishing defines a process that is about to finish.
const PoolProcessRequestFinishing string = "Finishing"

// PoolProcessRequestReadingHeaders defines a process that is reading headers.
const PoolProcessRequestReadingHeaders string = "Reading headers"

// PoolProcessRequestInfo defines a process that is getting request information.
const PoolProcessRequestInfo string = "Getting request informations"

// PoolProcessRequestEnding defines a process that is about to end.
const PoolProcessRequestEnding string = "Ending"

var log logger

type logger interface {
	Info(ar ...interface{})
	Infof(string, ...interface{})
	Debug(ar ...interface{})
	Debugf(string, ...interface{})
	Error(ar ...interface{})
	Errorf(string, ...interface{})
}

// PoolManager manages all configured Pools
type PoolManager struct {
	Pools []Pool `json:"pools"`
}

// Pool describes a single PHP-FPM pool that can be reached via a Socket or TCP address
type Pool struct {
	// The address of the pool, e.g. tcp://127.0.0.1:9000 or unix:///tmp/php-fpm.sock
	Address             string        `json:"-"`
	ScrapeError         error         `json:"-"`
	ScrapeFailures      int64         `json:"-"`
	Name                string        `json:"pool"`
	ProcessManager      string        `json:"process manager"`
	StartTime           timestamp     `json:"start time"`
	StartSince          int64         `json:"start since"`
	AcceptedConnections int64         `json:"accepted conn"`
	ListenQueue         int64         `json:"listen queue"`
	MaxListenQueue      int64         `json:"max listen queue"`
	ListenQueueLength   int64         `json:"listen queue len"`
	IdleProcesses       int64         `json:"idle processes"`
	ActiveProcesses     int64         `json:"active processes"`
	TotalProcesses      int64         `json:"total processes"`
	MaxActiveProcesses  int64         `json:"max active processes"`
	MaxChildrenReached  int64         `json:"max children reached"`
	SlowRequests        int64         `json:"slow requests"`
	Processes           []PoolProcess `json:"processes"`
}

type requestDuration int64

// PoolProcess describes a single PHP-FPM process. A pool can have multiple processes.
type PoolProcess struct {
	PID               int64           `json:"pid"`
	State             string          `json:"state"`
	StartTime         int64           `json:"start time"`
	StartSince        int64           `json:"start since"`
	Requests          int64           `json:"requests"`
	RequestDuration   requestDuration `json:"request duration"`
	RequestMethod     string          `json:"request method"`
	RequestURI        string          `json:"request uri"`
	ContentLength     int64           `json:"content length"`
	User              string          `json:"user"`
	Script            string          `json:"script"`
	LastRequestCPU    float64         `json:"last request cpu"`
	LastRequestMemory int64           `json:"last request memory"`
}

// PoolProcessStateCounter holds the calculated metrics for pool processes.
type PoolProcessStateCounter struct {
	Running        int64
	Idle           int64
	Finishing      int64
	ReadingHeaders int64
	Info           int64
	Ending         int64
}

// Add will add a pool to the pool manager based on the given URI.
func (pm *PoolManager) Add(uri string) Pool {
	p := Pool{Address: uri}
	pm.Pools = append(pm.Pools, p)
	return p
}

// Update will run the pool.Update() method concurrently on all Pools.
func (pm *PoolManager) Update() (err error) {
	wg := &sync.WaitGroup{}

	started := time.Now()

	for idx := range pm.Pools {
		wg.Add(1)
		go func(p *Pool) {
			defer wg.Done()
			if err := p.Update(); err != nil {
				log.Error(err)
			}
		}(&pm.Pools[idx])
	}

	wg.Wait()

	ended := time.Now()

	log.Debugf("Updated %v pool(s) in %v", len(pm.Pools), ended.Sub(started))

	return nil
}

// Update will connect to PHP-FPM and retrieve the latest data for the pool.
func (p *Pool) Update() (err error) {
	p.ScrapeError = nil

	scheme, address, path, err := parseURL(p.Address)
	if err != nil {
		return p.error(err)
	}

	fcgi, err := fcgiclient.DialTimeout(scheme, address, time.Duration(3)*time.Second)
	if err != nil {
		return p.error(err)
	}

	defer fcgi.Close()

	env := map[string]string{
		"SCRIPT_FILENAME": path,
		"SCRIPT_NAME":     path,
		"SERVER_SOFTWARE": "go / php-fpm_exporter",
		"REMOTE_ADDR":     "127.0.0.1",
		"QUERY_STRING":    "json&full",
	}

	resp, err := fcgi.Get(env)
	if err != nil {
		return p.error(err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.error(err)
	}

	content = JSONResponseFixer(content)

	log.Debugf("Pool[%v]: %v", p.Address, string(content))

	if err = json.Unmarshal(content, &p); err != nil {
		log.Errorf("Pool[%v]: %v", p.Address, string(content))
		return p.error(err)
	}

	return nil
}

func (p *Pool) error(err error) error {
	p.ScrapeError = err
	p.ScrapeFailures++
	log.Error(err)
	return err
}

// JSONResponseFixer resolves encoding issues with PHP-FPMs JSON response
func JSONResponseFixer(content []byte) []byte {
	c := string(content)
	re := regexp.MustCompile(`(,"request uri":)"(.*?)"(,"content length":)`)
	matches := re.FindAllStringSubmatch(c, -1)

	for _, match := range matches {
		requestURI, _ := json.Marshal(match[2])

		sold := match[0]
		snew := match[1] + string(requestURI) + match[3]

		c = strings.ReplaceAll(c, sold, snew)
	}

	return []byte(c)
}

// CountProcessState return the calculated metrics based on the reported processes.
func CountProcessState(processes []PoolProcess) (active int64, idle int64, total int64) {
	for idx := range processes {
		switch processes[idx].State {
		case PoolProcessRequestRunning:
			active++
		case PoolProcessRequestIdle:
			idle++
		case PoolProcessRequestEnding:
		case PoolProcessRequestFinishing:
		case PoolProcessRequestInfo:
		case PoolProcessRequestReadingHeaders:
			active++
		default:
			log.Errorf("Unknown process state '%v'", processes[idx].State)
		}
	}

	return active, idle, active + idle
}

// parseURL creates elements to be passed into fcgiclient.DialTimeout
func parseURL(rawurl string) (scheme string, address string, path string, err error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return uri.Scheme, uri.Host, uri.Path, err
	}

	scheme = uri.Scheme

	switch uri.Scheme {
	case "unix":
		result := strings.Split(uri.Path, ";")
		address = result[0]
		if len(result) > 1 {
			path = result[1]
		}
	default:
		address = uri.Host
		path = uri.Path
	}

	return
}

type timestamp time.Time

// MarshalJSON customise JSON for timestamp
func (t *timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

// UnmarshalJSON customise JSON for timestamp
func (t *timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*t = timestamp(time.Unix(int64(ts), 0))
	return nil
}

// This is because of bug in php-fpm that can return 'request duration' which can't
// fit to int64. For details check links:
// https://bugs.php.net/bug.php?id=62382
// https://serverfault.com/questions/624977/huge-request-duration-value-for-a-particular-php-script
func (rd *requestDuration) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprint(rd)
	return []byte(stamp), nil
}

func (rd *requestDuration) UnmarshalJSON(b []byte) error {
	rdc, err := strconv.Atoi(string(b))
	if err != nil {
		*rd = 0
	} else {
		*rd = requestDuration(rdc)
	}
	return nil
}

// SetLogger configures the used logger
func SetLogger(logger logger) {
	log = logger
}
