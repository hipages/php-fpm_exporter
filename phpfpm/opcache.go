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
)

// Opcache stores the opcache runtime statistics
type Opcache struct {
	OpcacheScriptPath string `json:"-"`

	OpcacheEnabled    bool `json:"opcache_enabled"`
	CacheFull         bool `json:"cache_full"`
	RestartPending    bool `json:"restart_pending"`
	RestartInProgress bool `json:"restart_in_progress"`
	MemoryUsage       struct {
		UsedMemory              int     `json:"used_memory"`
		FreeMemory              int     `json:"free_memory"`
		WastedMemory            int     `json:"wasted_memory"`
		CurrentWastedPercentage float64 `json:"current_wasted_percentage"`
	} `json:"memory_usage"`
	InternedStringsUsage struct {
		BufferSize      int `json:"buffer_size"`
		UsedMemory      int `json:"used_memory"`
		FreeMemory      int `json:"free_memory"`
		NumberOfStrings int `json:"number_of_strings"`
	} `json:"interned_strings_usage"`
	OpcacheStatistics struct {
		NumCachedScripts   int     `json:"num_cached_scripts"`
		NumCachedKeys      int     `json:"num_cached_keys"`
		MaxCachedKeys      int     `json:"max_cached_keys"`
		Hits               int     `json:"hits"`
		StartTime          int     `json:"start_time"`
		LastRestartTime    int     `json:"last_restart_time"`
		OomRestarts        int     `json:"oom_restarts"`
		HashRestarts       int     `json:"hash_restarts"`
		ManualRestarts     int     `json:"manual_restarts"`
		Misses             int     `json:"misses"`
		BlacklistMisses    int     `json:"blacklist_misses"`
		BlacklistMissRatio int     `json:"blacklist_miss_ratio"`
		OpcacheHitRate     float64 `json:"opcache_hit_rate"`
	} `json:"opcache_statistics"`
}

// Update the opcache statistics
func (oc *Opcache) Update(p Pool) (err error) {
	content, err := p.Execute(oc.OpcacheScriptPath)
	if err != nil {
		return p.error(err)
	}

	if err = json.Unmarshal(content, &oc); err != nil {
		log.Errorf("Pool[%v]: %v", p.Name, string(content))
		return p.error(err)
	}

	log.Debugf("Opcache statistics returned %v", oc)
	return nil
}
