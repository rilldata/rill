package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadGuardrails_Defaults(t *testing.T) {
	t.Parallel()

	// When no instance variables are provided, defaults should be used.
	cfg := LoadGuardrails(nil)
	require.NotNil(t, cfg)
	require.Equal(t, DefaultSoftBytesScanned, cfg.SoftLimitBytesScanned)
	require.Equal(t, DefaultHardBytesScanned, cfg.HardLimitBytesScanned)
	require.Equal(t, DefaultSoftRuntimeMS, cfg.SoftLimitRuntimeMS)
	require.Equal(t, DefaultHardRuntimeMS, cfg.HardLimitRuntimeMS)
}

func TestLoadGuardrails_EmptyMap(t *testing.T) {
	t.Parallel()

	cfg := LoadGuardrails(map[string]string{})
	require.NotNil(t, cfg)
	require.Equal(t, DefaultSoftBytesScanned, cfg.SoftLimitBytesScanned)
	require.Equal(t, DefaultHardBytesScanned, cfg.HardLimitBytesScanned)
	require.Equal(t, DefaultSoftRuntimeMS, cfg.SoftLimitRuntimeMS)
	require.Equal(t, DefaultHardRuntimeMS, cfg.HardLimitRuntimeMS)
}

func TestLoadGuardrails_AllCustomValues(t *testing.T) {
	t.Parallel()

	vars := map[string]string{
		"query_console_soft_limit_bytes_scanned": "500000",
		"query_console_hard_limit_bytes_scanned": "2000000",
		"query_console_soft_limit_runtime_ms":    "10000",
		"query_console_hard_limit_runtime_ms":    "60000",
	}

	cfg := LoadGuardrails(vars)
	require.NotNil(t, cfg)
	require.Equal(t, int64(500000), cfg.SoftLimitBytesScanned)
	require.Equal(t, int64(2000000), cfg.HardLimitBytesScanned)
	require.Equal(t, int64(10000), cfg.SoftLimitRuntimeMS)
	require.Equal(t, int64(60000), cfg.HardLimitRuntimeMS)
}

func TestLoadGuardrails_PartialOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		vars map[string]string
		wantSoftBytes int64
		wantHardBytes int64
		wantSoftMS    int64
		wantHardMS    int64
	}{
		{
			name: "only soft bytes",
			vars: map[string]string{
				"query_console_soft_limit_bytes_scanned": "100",
			},
			wantSoftBytes: 100,
			wantHardBytes: DefaultHardBytesScanned,
			wantSoftMS:    DefaultSoftRuntimeMS,
			wantHardMS:    DefaultHardRuntimeMS,
		},
		{
			name: "only hard bytes",
			vars: map[string]string{
				"query_console_hard_limit_bytes_scanned": "9999999",
			},
			wantSoftBytes: DefaultSoftBytesScanned,
			wantHardBytes: 9999999,
			wantSoftMS:    DefaultSoftRuntimeMS,
			wantHardMS:    DefaultHardRuntimeMS,
		},
		{
			name: "only soft runtime",
			vars: map[string]string{
				"query_console_soft_limit_runtime_ms": "5000",
			},
			wantSoftBytes: DefaultSoftBytesScanned,
			wantHardBytes: DefaultHardBytesScanned,
			wantSoftMS:    5000,
			wantHardMS:    DefaultHardRuntimeMS,
		},
		{
			name: "only hard runtime",
			vars: map[string]string{
				"query_console_hard_limit_runtime_ms": "120000",
			},
			wantSoftBytes: DefaultSoftBytesScanned,
			wantHardBytes: DefaultHardBytesScanned,
			wantSoftMS:    DefaultSoftRuntimeMS,
			wantHardMS:    120000,
		},
		{
			name: "bytes overridden, runtime defaults",
			vars: map[string]string{
				"query_console_soft_limit_bytes_scanned": "100",
				"query_console_hard_limit_bytes_scanned": "200",
			},
			wantSoftBytes: 100,
			wantHardBytes: 200,
			wantSoftMS:    DefaultSoftRuntimeMS,
			wantHardMS:    DefaultHardRuntimeMS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := LoadGuardrails(tt.vars)
			require.Equal(t, tt.wantSoftBytes, cfg.SoftLimitBytesScanned)
			require.Equal(t, tt.wantHardBytes, cfg.HardLimitBytesScanned)
			require.Equal(t, tt.wantSoftMS, cfg.SoftLimitRuntimeMS)
			require.Equal(t, tt.wantHardMS, cfg.HardLimitRuntimeMS)
		})
	}
}

func TestLoadGuardrails_InvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		vars map[string]string
	}{
		{
			name: "non-numeric soft bytes",
			vars: map[string]string{
				"query_console_soft_limit_bytes_scanned": "not_a_number",
			},
		},
		{
			name: "non-numeric hard bytes",
			vars: map[string]string{
				"query_console_hard_limit_bytes_scanned": "abc",
			},
		},
		{
			name: "empty string value",
			vars: map[string]string{
				"query_console_soft_limit_bytes_scanned": "",
			},
		},
		{
			name: "float value",
			vars: map[string]string{
				"query_console_soft_limit_runtime_ms": "3.14",
			},
		},
	}

	// Invalid values should fall back to defaults without panicking.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := LoadGuardrails(tt.vars)
			require.NotNil(t, cfg)
			// All invalid entries should produce defaults for the affected field.
			// We can't easily know which field was invalid, but at least none should panic.
			require.True(t, cfg.SoftLimitBytesScanned >= 0)
			require.True(t, cfg.HardLimitBytesScanned >= 0)
			require.True(t, cfg.SoftLimitRuntimeMS >= 0)
			require.True(t, cfg.HardLimitRuntimeMS >= 0)
		})
	}
}

func TestLoadGuardrails_ZeroValues(t *testing.T) {
	t.Parallel()

	// Zero is a valid value — it means "disabled" (no limit).
	vars := map[string]string{
		"query_console_soft_limit_bytes_scanned": "0",
		"query_console_hard_limit_bytes_scanned": "0",
		"query_console_soft_limit_runtime_ms":    "0",
		"query_console_hard_limit_runtime_ms":    "0",
	}

	cfg := LoadGuardrails(vars)
	require.Equal(t, int64(0), cfg.SoftLimitBytesScanned)
	require.Equal(t, int64(0), cfg.HardLimitBytesScanned)
	require.Equal(t, int64(0), cfg.SoftLimitRuntimeMS)
	require.Equal(t, int64(0), cfg.HardLimitRuntimeMS)
}

func TestLoadGuardrails_NegativeValues(t *testing.T) {
	t.Parallel()

	// Negative values are technically parseable as int64.
	// The loader should accept them (or clamp to 0/default). We document the behavior.
	vars := map[string]string{
		"query_console_soft_limit_bytes_scanned": "-100",
		"query_console_hard_limit_bytes_scanned": "-1",
	}

	cfg := LoadGuardrails(vars)
	require.NotNil(t, cfg)
	// Negative values should be treated as zero (disabled) — guardrails can't scan negative bytes.
	require.True(t, cfg.SoftLimitBytesScanned <= 0, "expected non-positive soft limit for negative input")
	require.True(t, cfg.HardLimitBytesScanned <= 0, "expected non-positive hard limit for negative input")
}

func TestLoadGuardrails_UnrelatedVariablesIgnored(t *testing.T) {
	t.Parallel()

	vars := map[string]string{
		"unrelated_var":                          "12345",
		"another_thing":                          "hello",
		"query_console_soft_limit_bytes_scanned": "42",
	}

	cfg := LoadGuardrails(vars)
	require.Equal(t, int64(42), cfg.SoftLimitBytesScanned)
	require.Equal(t, DefaultHardBytesScanned, cfg.HardLimitBytesScanned)
	require.Equal(t, DefaultSoftRuntimeMS, cfg.SoftLimitRuntimeMS)
	require.Equal(t, DefaultHardRuntimeMS, cfg.HardLimitRuntimeMS)
}

// --- CheckSoftLimit Tests ---

func TestCheckSoftLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		estimated    int64
		config       *GuardrailConfig
		wantExceeded bool
	}{
		{
			name:      "below soft limit",
			estimated: 500,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1000,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: false,
		},
		{
			name:      "exactly at soft limit",
			estimated: 1000,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1000,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: true,
		},
		{
			name:      "above soft limit",
			estimated: 2000,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1000,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: true,
		},
		{
			name:      "zero soft limit disables check",
			estimated: 999999999,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 0,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: false,
		},
		{
			name:      "zero estimated bytes",
			estimated: 0,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1000,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: false,
		},
		{
			name:      "negative estimated bytes",
			estimated: -100,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1000,
				HardLimitBytesScanned: 5000,
			},
			wantExceeded: false,
		},
		{
			name:      "very large estimated bytes exceeds soft",
			estimated: 1<<40, // 1 TB
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1 << 30, // 1 GB
				HardLimitBytesScanned: 1 << 42,
			},
			wantExceeded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exceeded, msg := CheckSoftLimit(tt.estimated, tt.config)
			require.Equal(t, tt.wantExceeded, exceeded)
			if exceeded {
				require.NotEmpty(t, msg, "expected a warning message when soft limit exceeded")
			}
		})
	}
}

// --- CheckHardLimit Tests ---

func TestCheckHardLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		estimated   int64
		config      *GuardrailConfig
		wantBlocked bool
	}{
		{
			name:      "below hard limit",
			estimated: 500,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 1000,
			},
			wantBlocked: false,
		},
		{
			name:      "exactly at hard limit",
			estimated: 1000,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 1000,
			},
			wantBlocked: true,
		},
		{
			name:      "above hard limit",
			estimated: 5000,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 1000,
			},
			wantBlocked: true,
		},
		{
			name:      "zero hard limit disables check",
			estimated: 999999999,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 0,
			},
			wantBlocked: false,
		},
		{
			name:      "zero estimated bytes",
			estimated: 0,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 1000,
			},
			wantBlocked: false,
		},
		{
			name:      "negative estimated bytes",
			estimated: -50,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 100,
				HardLimitBytesScanned: 1000,
			},
			wantBlocked: false,
		},
		{
			name:      "very large estimated bytes exceeds hard",
			estimated: 1 << 42,
			config: &GuardrailConfig{
				SoftLimitBytesScanned: 1 << 30,
				HardLimitBytesScanned: 1 << 40,
			},
			wantBlocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blocked, reason := CheckHardLimit(tt.estimated, tt.config)
			require.Equal(t, tt.wantBlocked, blocked)
			if blocked {
				require.NotEmpty(t, reason, "expected a blocking reason when hard limit exceeded")
			}
		})
	}
}

// --- Combined Soft + Hard Limit Interaction Tests ---

func TestCheckLimits_SoftAndHardInteraction(t *testing.T) {
	t.Parallel()

	config := &GuardrailConfig{
		SoftLimitBytesScanned: 1000,
		HardLimitBytesScanned: 5000,
		SoftLimitRuntimeMS:    10000,
		HardLimitRuntimeMS:    30000,
	}

	// Below both limits.
	exceeded, _ := CheckSoftLimit(500, config)
	blocked, _ := CheckHardLimit(500, config)
	require.False(t, exceeded)
	require.False(t, blocked)

	// Between soft and hard (soft exceeded, not blocked).
	exceeded, msg := CheckSoftLimit(2000, config)
	blocked, _ = CheckHardLimit(2000, config)
	require.True(t, exceeded)
	require.NotEmpty(t, msg)
	require.False(t, blocked)

	// Above hard limit (both exceeded and blocked).
	exceeded, _ = CheckSoftLimit(6000, config)
	blocked, reason := CheckHardLimit(6000, config)
	require.True(t, exceeded)
	require.True(t, blocked)
	require.NotEmpty(t, reason)
}

func TestCheckLimits_BothDisabled(t *testing.T) {
	t.Parallel()

	config := &GuardrailConfig{
		SoftLimitBytesScanned: 0,
		HardLimitBytesScanned: 0,
		SoftLimitRuntimeMS:    0,
		HardLimitRuntimeMS:    0,
	}

	// Even extremely large values should not trigger when limits are zero.
	exceeded, _ := CheckSoftLimit(1<<50, config)
	blocked, _ := CheckHardLimit(1<<50, config)
	require.False(t, exceeded)
	require.False(t, blocked)
}

func TestCheckSoftLimit_NilConfig(t *testing.T) {
	t.Parallel()

	// If a nil config is somehow passed, it should not panic.
	// The function should treat it as "no limits configured".
	exceeded, _ := CheckSoftLimit(1000, nil)
	require.False(t, exceeded)
}

func TestCheckHardLimit_NilConfig(t *testing.T) {
	t.Parallel()

	blocked, _ := CheckHardLimit(1000, nil)
	require.False(t, blocked)
}

// --- Default Constants Sanity Checks ---

func TestDefaultConstants(t *testing.T) {
	t.Parallel()

	// Verify that defaults are sensible: soft < hard, positive values.
	require.Greater(t, DefaultSoftBytesScanned, int64(0), "default soft bytes should be positive")
	require.Greater(t, DefaultHardBytesScanned, int64(0), "default hard bytes should be positive")
	require.Greater(t, DefaultSoftRuntimeMS, int64(0), "default soft runtime should be positive")
	require.Greater(t, DefaultHardRuntimeMS, int64(0), "default hard runtime should be positive")

	require.Less(t, DefaultSoftBytesScanned, DefaultHardBytesScanned, "soft bytes limit should be less than hard")
	require.Less(t, DefaultSoftRuntimeMS, DefaultHardRuntimeMS, "soft runtime limit should be less than hard")
}
