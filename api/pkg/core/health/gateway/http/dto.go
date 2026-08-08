// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"

type statusResponse struct {
	Status string `json:"status"`
}

type componentResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type readyzResponse struct {
	Components map[string]componentResponse `json:"components"`
	Status     string                       `json:"status"`
}

const probeFailureReason = "unreachable"

func newReadyzResponse(rep health.Report) readyzResponse {
	components := make(map[string]componentResponse, len(rep.Components))
	for name, c := range rep.Components {
		components[name] = componentResponse{Status: string(c.Status), Reason: publicReason(c)}
	}

	return readyzResponse{Status: string(rep.Status), Components: components}
}

func publicReason(c health.Component) string {
	if c.Probe && c.Reason != "" {
		return probeFailureReason
	}

	return c.Reason
}
