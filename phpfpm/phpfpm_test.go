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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountProcessState(t *testing.T) {
	processes := []PoolProcess{
		{State: PoolProcessRequestIdle},
		{State: PoolProcessRequestRunning},
		{State: PoolProcessRequestReadingHeaders},
		{State: PoolProcessRequestInfo},
		{State: PoolProcessRequestFinishing},
		{State: PoolProcessRequestEnding},
	}

	active, idle, total := CountProcessState(processes)

	assert.Equal(t, int64(2), active, "active processes")
	assert.Equal(t, int64(1), idle, "idle processes")
	assert.Equal(t, int64(3), total, "total processes")
}

// https://github.com/hipages/php-fpm_exporter/issues/10
func TestCannotUnmarshalNumberIssue10(t *testing.T) {
	pool := Pool{}
	content := []byte(`{
	   "pool":"www",
	   "process manager":"dynamic",
	   "start time":1519474655,
	   "start since":302035,
	   "accepted conn":44144,
	   "listen queue":0,
	   "max listen queue":1,
	   "listen queue len":128,
	   "idle processes":1,
	   "active processes":1,
	   "total processes":2,
	   "max active processes":2,
	   "max children reached":0,
	   "slow requests":0,
	   "processes":[
		  {
			 "pid":23,
			 "state":"Idle",
			 "start time":1519474655,
			 "start since":302035,
			 "requests":22071,
			 "request duration":295,
			 "request method":"GET",
			 "request uri":"/status?json&full",
			 "content length":0,
			 "user":"-",
			 "script":"-",
			 "last request cpu":0.00,
			 "last request memory":2097152
		  },
		  {
			 "pid":24,
			 "state":"Running",
			 "start time":1519474655,
			 "start since":302035,
			 "requests":22073,
			 "request duration":18446744073709550774,
			 "request method":"GET",
			 "request uri":"/status?json&full",
			 "content length":0,
			 "user":"-",
			 "script":"-",
			 "last request cpu":0.00,
			 "last request memory":0
		  }
	   ]
    }`)

	err := json.Unmarshal(content, &pool)

	assert.NotNil(t, err, err.Error())
}

func TestParseURL(t *testing.T) {
	var uris = []struct {
		in  string
		out []string
		err error
	}{
		{"tcp://127.0.0.1:9000/status", []string{"tcp", "127.0.0.1:9000", "/status"}, nil},
		{"tcp://127.0.0.1", []string{"tcp", "127.0.0.1", ""}, nil},
		{"unix:///tmp/php.sock;/status", []string{"unix", "/tmp/php.sock", "/status"}, nil},
		{"unix:///tmp/php.sock", []string{"unix", "/tmp/php.sock", ""}, nil},
	}

	for _, u := range uris {
		scheme, address, path, err := parseURL(u.in)
		assert.Equal(t, u.err, err)
		assert.Equal(t, u.out, []string{scheme, address, path})
	}
}
