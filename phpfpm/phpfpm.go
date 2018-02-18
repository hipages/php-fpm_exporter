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

package phpfpm

import (
	"encoding/json"
	"fmt"
	"github.com/tomasen/fcgi_client"
	"io/ioutil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var log logger

type logger interface {
	Debugf(string, ...interface{})
	Error(ar ...interface{})
}

type PoolManager struct {
	pools []Pool `json:"pools"`
}

type Pool struct {
	// The address of the pool, e.g. tcp://127.0.0.1:9000 or unix:///tmp/php-fpm.sock
	Address             string
	CollectionError     error
	Name                string        `json:"pool"`
	ProcessManager      string        `json:"process manager"`
	StartTime           Timestamp     `json:"start time"`
	StartSince          int           `json:"start since"`
	AcceptedConnections int           `json:"accepted conn"`
	ListenQueue         int           `json:"listen queue"`
	MaxListenQueue      int           `json:"max listen queue"`
	ListenQueueLength   int           `json:"listen queue len"`
	IdleProcesses       int           `json:"idle processes"`
	ActiveProcesses     int           `json:"active processes"`
	TotalProcesses      int           `json:"total processes"`
	MaxActiveProcesses  int           `json:"max active processes"`
	MaxChildrenReached  int           `json:"max children reached"`
	SlowRequests        int           `json:"slow requests"`
	Processes           []PoolProcess `json:"processes"`
}

type PoolProcess struct {
	PID               int     `json:"pid"`
	State             string  `json:"state"`
	StartTime         int     `json:"start time"`
	StartSince        int     `json:"start since"`
	Requests          int     `json:"requests"`
	RequestDuration   int     `json:"request duration"`
	RequestMethod     string  `json:"request method"`
	RequestURI        string  `json:"request uri"`
	ContentLength     int     `json:"content length"`
	User              string  `json:"user"`
	Script            string  `json:"script"`
	LastRequestCPU    float32 `json:"last request cpu"`
	LastRequestMemory int     `json:"last request memory"`
}

func (pm *PoolManager) Add(uri string) Pool {
	p := Pool{Address: uri}
	pm.pools = append(pm.pools, p)
	return p
}

func (pm *PoolManager) Update() (err error) {
	wg := &sync.WaitGroup{}

	started := time.Now()

	for idx := range pm.pools {
		wg.Add(1)
		go func(p *Pool) {
			defer wg.Done()
			p.Update()
		}(&pm.pools[idx])
	}

	wg.Wait()

	ended := time.Now()

	log.Debugf("Updated %v pool(s) in %v", len(pm.pools), ended.Sub(started))

	return nil
}

func (pm *PoolManager) Pools() []Pool {
	return pm.pools
}

// Implement custom Marshaler due to "pools" being unexported
func (pm PoolManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Pools []Pool `json:"pools"`
	}{Pools: pm.pools})
}

func (p *Pool) Update() (err error) {
	p.CollectionError = nil

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = "/status"
	env["SCRIPT_NAME"] = "/status"
	env["SERVER_SOFTWARE"] = "go / php-fpm_exporter "
	env["REMOTE_ADDR"] = "127.0.0.1"
	env["QUERY_STRING"] = "json&full"

	uri, err := url.Parse(p.Address)
	if err != nil {
		return p.error(err)
	}

	fcgi, err := fcgiclient.DialTimeout(uri.Scheme, uri.Hostname()+":"+uri.Port(), time.Duration(3)*time.Second)
	if err != nil {
		return p.error(err)
	}

	resp, err := fcgi.Get(env)
	if err != nil {
		return p.error(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.error(err)
	}

	fcgi.Close()

	log.Debugf("Pool[", p.Address, "]:", string(content))

	if err = json.Unmarshal(content, &p); err != nil {
		return p.error(err)
	}

	return nil
}

func (p *Pool) error(err error) error {
	p.CollectionError = err
	log.Error(err)
	return err
}

type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(int64(ts), 0))
	return nil
}

func SetLogger(logger logger) {
	log = logger
}
