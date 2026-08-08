// SPDX-License-Identifier: AGPL-3.0-or-later

package health

import (
	"context"
	"sync"
)

type Status string

const (
	StatusStarting Status = "starting"
	StatusOK       Status = "ok"
	StatusFailed   Status = "failed"
)

type Probe func(context.Context) error

type Component struct {
	Status Status
	Reason string
	Probe  bool
}

type Report struct {
	Components map[string]Component
	Status     Status
}

type Registry struct {
	latches map[string]Component
	probes  map[string]Probe
	mtx     sync.RWMutex
}

func NewRegistry(latches ...string) *Registry {
	r := &Registry{
		latches: make(map[string]Component, len(latches)),
		probes:  make(map[string]Probe),
	}

	for _, name := range latches {
		r.latches[name] = Component{Status: StatusStarting}
	}

	return r
}

func (r *Registry) AddProbe(name string, p Probe) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.probes[name] = p
}

func (r *Registry) Set(name string, err error) {
	c := Component{Status: StatusOK}
	if err != nil {
		c = Component{Status: StatusFailed, Reason: err.Error()}
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.latches[name] = c
}

func (r *Registry) LatchStatus(name string) Status {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	c, ok := r.latches[name]
	if !ok {
		return StatusStarting
	}

	return c.Status
}

func (r *Registry) Snapshot(ctx context.Context) Report {
	r.mtx.RLock()

	components := make(map[string]Component, len(r.latches)+len(r.probes))
	for name, c := range r.latches {
		components[name] = c
	}

	probes := make(map[string]Probe, len(r.probes))
	for name, p := range r.probes {
		probes[name] = p
	}

	r.mtx.RUnlock()

	for name, p := range probes {
		if err := p(ctx); err != nil {
			components[name] = Component{Status: StatusFailed, Reason: err.Error(), Probe: true}

			continue
		}

		components[name] = Component{Status: StatusOK, Probe: true}
	}

	return Report{Status: aggregate(components), Components: components}
}

func aggregate(components map[string]Component) Status {
	status := StatusOK

	for _, c := range components {
		switch c.Status {
		case StatusFailed:
			return StatusFailed
		case StatusStarting:
			status = StatusStarting
		case StatusOK:
		}
	}

	return status
}
