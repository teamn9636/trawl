package extract

// HealthStats tracks extraction quality across a batch of pages.
type HealthStats struct {
	TotalRecords    int
	TotalFields     int
	PopulatedFields int
	EmptyFields     int
}

// SuccessRate returns the percentage of fields that were populated.
func (h *HealthStats) SuccessRate() float64 {
	if h.TotalFields == 0 {
		return 0
	}
	return float64(h.PopulatedFields) / float64(h.TotalFields) * 100
}

// NeedsReInference returns true if the success rate is below the threshold.
func (h *HealthStats) NeedsReInference(threshold float64) bool {
	return h.SuccessRate() < threshold
}

// ComputeHealth analyzes an extraction result and returns health stats.
func ComputeHealth(result *Result) *HealthStats {
	stats := &HealthStats{
		TotalRecords: len(result.Records),
	}

	for _, rec := range result.Records {
		for _, field := range result.Fields {
			stats.TotalFields++
			val, ok := rec[field]
			if ok && val != nil && val != "" {
				stats.PopulatedFields++
			} else {
				stats.EmptyFields++
			}
		}
	}

	return stats
}
