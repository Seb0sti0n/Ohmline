package engine

import (
	"math"
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

// Flag kinds raised on a single reading.
const (
	FlagVoltageRange = "voltage_out_of_range"
	FlagVoltageJump  = "voltage_jump"
	FlagPFJump       = "power_factor_jump"
	FlagCoherence    = "kwh_incoherent_with_v_i_pf"
)

type QualityInfo struct {
	TotalFlags  int            `json:"total_flags"`
	ByKind      map[string]int `json:"by_kind"`
	Recurrent   bool           `json:"recurrent"`
	FirstFlag   time.Time      `json:"first_flag"`
	LastFlag    time.Time      `json:"last_flag"`
	FlagsIn24h  int            `json:"max_flags_in_24h"`
	FlagHours   int            `json:"flagged_hours"`
	WindowHours int            `json:"window_hours"`
}

// QualityResult is the outcome of the data-quality scan of one meter. StartIdx/EndIdx delimit the
// window from the first flag of the first recurrent cluster to the last flag; both are -1 if none.
type QualityResult struct {
	Info             QualityInfo
	StartIdx, EndIdx int
	Flagged          []int // reading indexes with at least one flag
}

// DetectQuality flags readings whose electrical values are implausible or inconsistent.
// The status column of the source data is deliberately ignored (it is always "OK").
func DetectQuality(rs []Reading, b Baseline, cfg config.Engine) QualityResult {
	res := QualityResult{StartIdx: -1, EndIdx: -1, Info: QualityInfo{ByKind: map[string]int{}}}
	tol := cfg.VoltageNominal * cfg.VoltageTolerancePct / 100

	for i, r := range rs {
		var kinds []string
		if math.Abs(r.VoltageV-cfg.VoltageNominal) > tol {
			kinds = append(kinds, FlagVoltageRange)
		}
		if i > 0 {
			if math.Abs(r.VoltageV-rs[i-1].VoltageV) > cfg.VoltageJumpV {
				kinds = append(kinds, FlagVoltageJump)
			}
			if math.Abs(r.PowerFactor-rs[i-1].PowerFactor) > cfg.PFJump {
				kinds = append(kinds, FlagPFJump)
			}
		}
		if e := electricalKWh(r); e > 0 && b.CoherenceRatio > 0 {
			if math.Abs((r.ConsumptionKWh/e)/b.CoherenceRatio-1)*100 > cfg.CoherenceTolPct {
				kinds = append(kinds, FlagCoherence)
			}
		}
		if len(kinds) > 0 {
			res.Flagged = append(res.Flagged, i)
			for _, k := range kinds {
				res.Info.ByKind[k]++
				res.Info.TotalFlags++
			}
		}
	}
	res.Info.FlagHours = len(res.Flagged)
	if len(res.Flagged) == 0 {
		return res
	}

	// Recurrent = at least QualityFlagsPer24h flagged hours within a 24 h span.
	n := cfg.QualityFlagsPer24h
	for a := 0; a+n-1 < len(res.Flagged); a++ {
		span := rs[res.Flagged[a+n-1]].Timestamp.Sub(rs[res.Flagged[a]].Timestamp)
		if span < 24*time.Hour {
			if !res.Info.Recurrent {
				res.Info.Recurrent = true
				res.StartIdx = res.Flagged[a]
			}
		}
	}
	for a := range res.Flagged {
		cnt := 0
		for c := a; c < len(res.Flagged) && rs[res.Flagged[c]].Timestamp.Sub(rs[res.Flagged[a]].Timestamp) < 24*time.Hour; c++ {
			cnt++
		}
		if cnt > res.Info.FlagsIn24h {
			res.Info.FlagsIn24h = cnt
		}
	}
	res.Info.FirstFlag = rs[res.Flagged[0]].Timestamp
	res.Info.LastFlag = rs[res.Flagged[len(res.Flagged)-1]].Timestamp
	if res.Info.Recurrent {
		res.EndIdx = res.Flagged[len(res.Flagged)-1]
		res.Info.WindowHours = res.EndIdx - res.StartIdx + 1
	}
	return res
}
