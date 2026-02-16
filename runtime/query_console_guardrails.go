package runtime

import (
	"fmt"
	"strconv"
)

// Default guardrail thresholds for the query console.
const (
	// DefaultSoftBytesScanned is the default soft limit for bytes scanned (1 GB).
	DefaultSoftBytesScanned int64 = 1 << 30 // 1 GiB

	// DefaultHardBytesScanned is the default hard limit for bytes scanned (10 GB).
	DefaultHardBytesScanned int64 = 10 << 30 // 10 GiB

	// DefaultSoftRuntimeMS is the default soft limit for query runtime in milliseconds (30 seconds).
	DefaultSoftRuntimeMS int64 = 30_000

	// DefaultHardRuntimeMS is the default hard limit for query runtime in milliseconds (5 minutes).
	DefaultHardRuntimeMS int64 = 300_000
)

// Instance variable keys for guardrail configuration.
const (
	QueryConsoleSoftBytesScannedVar = "query_console.soft_bytes_scanned"
	QueryConsoleHardBytesScannedVar = "query_console.hard_bytes_scanned"
	QueryConsoleSoftRuntimeMSVar    = "query_console.soft_runtime_ms"
	QueryConsoleHardRuntimeMSVar    = "query_console.hard_runtime_ms"
)

// GuardrailConfig holds the soft and hard thresholds for query console guardrails.
// Soft limits produce a warning that the user can override.
// Hard limits block the query unconditionally.
type GuardrailConfig struct {
	// SoftBytesScanned is the soft limit for estimated bytes scanned.
	// When exceeded, the user receives a cost warning and may choose to proceed.
	SoftBytesScanned int64

	// HardBytesScanned is the hard limit for estimated bytes scanned.
	// When exceeded, the query is blocked entirely.
	HardBytesScanned int64

	// SoftRuntimeMS is the soft limit for estimated query runtime in milliseconds.
	// When exceeded, the user receives a warning and may choose to proceed.
	SoftRuntimeMS int64

	// HardRuntimeMS is the hard limit for query runtime in milliseconds.
	// When exceeded, the query is blocked entirely.
	HardRuntimeMS int64
}

// LoadGuardrails reads guardrail configuration from instance variables.
// Any missing or invalid values fall back to defaults.
func LoadGuardrails(instanceVariables map[string]string) *GuardrailConfig {
	cfg := &GuardrailConfig{
		SoftBytesScanned: DefaultSoftBytesScanned,
		HardBytesScanned: DefaultHardBytesScanned,
		SoftRuntimeMS:    DefaultSoftRuntimeMS,
		HardRuntimeMS:    DefaultHardRuntimeMS,
	}

	if instanceVariables == nil {
		return cfg
	}

	if v, ok := instanceVariables[QueryConsoleSoftBytesScannedVar]; ok {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			cfg.SoftBytesScanned = parsed
		}
	}

	if v, ok := instanceVariables[QueryConsoleHardBytesScannedVar]; ok {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			cfg.HardBytesScanned = parsed
		}
	}

	if v, ok := instanceVariables[QueryConsoleSoftRuntimeMSVar]; ok {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			cfg.SoftRuntimeMS = parsed
		}
	}

	if v, ok := instanceVariables[QueryConsoleHardRuntimeMSVar]; ok {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			cfg.HardRuntimeMS = parsed
		}
	}

	// Ensure hard limits are at least as large as soft limits.
	// If misconfigured, clamp hard to soft so that soft warnings always fire before a hard block.
	if cfg.HardBytesScanned < cfg.SoftBytesScanned {
		cfg.HardBytesScanned = cfg.SoftBytesScanned
	}
	if cfg.HardRuntimeMS < cfg.SoftRuntimeMS {
		cfg.HardRuntimeMS = cfg.SoftRuntimeMS
	}

	return cfg
}

// CheckSoftBytesLimit checks whether the estimated bytes scanned exceeds the soft limit.
// If exceeded, it returns true with a human-readable warning message.
func CheckSoftBytesLimit(estimatedBytes int64, config *GuardrailConfig) (exceeded bool, message string) {
	if config == nil || estimatedBytes <= 0 {
		return false, ""
	}
	if estimatedBytes > config.SoftBytesScanned {
		return true, fmt.Sprintf(
			"This query is estimated to scan %s, which exceeds the soft limit of %s. Do you want to proceed?",
			formatBytes(estimatedBytes),
			formatBytes(config.SoftBytesScanned),
		)
	}
	return false, ""
}

// CheckHardBytesLimit checks whether the estimated bytes scanned exceeds the hard limit.
// If exceeded, it returns true with a reason indicating the query is blocked.
func CheckHardBytesLimit(estimatedBytes int64, config *GuardrailConfig) (blocked bool, reason string) {
	if config == nil || estimatedBytes <= 0 {
		return false, ""
	}
	if estimatedBytes > config.HardBytesScanned {
		return true, fmt.Sprintf(
			"Query blocked: estimated scan of %s exceeds the hard limit of %s.",
			formatBytes(estimatedBytes),
			formatBytes(config.HardBytesScanned),
		)
	}
	return false, ""
}

// CheckSoftRuntimeLimit checks whether the estimated runtime exceeds the soft limit.
// If exceeded, it returns true with a human-readable warning message.
func CheckSoftRuntimeLimit(estimatedMS int64, config *GuardrailConfig) (exceeded bool, message string) {
	if config == nil || estimatedMS <= 0 {
		return false, ""
	}
	if estimatedMS > config.SoftRuntimeMS {
		return true, fmt.Sprintf(
			"This query is estimated to take %s, which exceeds the soft limit of %s. Do you want to proceed?",
			formatDurationMS(estimatedMS),
			formatDurationMS(config.SoftRuntimeMS),
		)
	}
	return false, ""
}

// CheckHardRuntimeLimit checks whether the estimated runtime exceeds the hard limit.
// If exceeded, it returns true with a reason indicating the query is blocked.
func CheckHardRuntimeLimit(estimatedMS int64, config *GuardrailConfig) (blocked bool, reason string) {
	if config == nil || estimatedMS <= 0 {
		return false, ""
	}
	if estimatedMS > config.HardRuntimeMS {
		return true, fmt.Sprintf(
			"Query blocked: estimated runtime of %s exceeds the hard limit of %s.",
			formatDurationMS(estimatedMS),
			formatDurationMS(config.HardRuntimeMS),
		)
	}
	return false, ""
}

// formatBytes returns a human-readable representation of a byte count.
func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)

	switch {
	case b >= tb:
		return fmt.Sprintf("%.2f TB", float64(b)/float64(tb))
	case b >= gb:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}

// formatDurationMS returns a human-readable representation of a duration given in milliseconds.
func formatDurationMS(ms int64) string {
	switch {
	case ms >= 60_000:
		minutes := float64(ms) / 60_000
		return fmt.Sprintf("%.1f minutes", minutes)
	case ms >= 1_000:
		seconds := float64(ms) / 1_000
		return fmt.Sprintf("%.1f seconds", seconds)
	default:
		return fmt.Sprintf("%d ms", ms)
	}
}
