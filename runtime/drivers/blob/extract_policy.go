package blob

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
)

type ExtractPolicy struct {
	RowsStrategy   ExtractPolicyStrategy
	RowsLimitBytes uint64
	FilesStrategy  ExtractPolicyStrategy
	FilesLimit     uint64
}

type ExtractPolicyStrategy int

const (
	ExtractPolicyStrategyUnspecified ExtractPolicyStrategy = 0
	ExtractPolicyStrategyHead        ExtractPolicyStrategy = 1
	ExtractPolicyStrategyTail        ExtractPolicyStrategy = 2
)

func (s ExtractPolicyStrategy) String() string {
	switch s {
	case ExtractPolicyStrategyHead:
		return "head"
	case ExtractPolicyStrategyTail:
		return "tail"
	default:
		return "unspecified"
	}
}

type rawExtractPolicy struct {
	Rows *struct {
		Strategy string `mapstructure:"strategy"`
		Size     string `mapstructure:"size"`
	} `mapstructure:"rows"`
	Files *struct {
		Strategy string `mapstructure:"strategy"`
		Size     string `mapstructure:"size"`
	} `mapstructure:"files"`
}

func ParseExtractPolicy(cfg map[string]any) (*ExtractPolicy, error) {
	if len(cfg) == 0 {
		return nil, nil
	}

	raw := &rawExtractPolicy{}
	if err := mapstructure.WeakDecode(cfg, raw); err != nil {
		return nil, err
	}

	res := &ExtractPolicy{}

	// Parse files
	if raw.Files != nil {
		strategy, err := parseStrategy(raw.Files.Strategy)
		if err != nil {
			return nil, err
		}
		res.FilesStrategy = strategy

		size, err := strconv.ParseUint(raw.Files.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}
		res.FilesLimit = size
	}

	// Parse rows
	if raw.Rows != nil {
		strategy, err := parseStrategy(raw.Rows.Strategy)
		if err != nil {
			return nil, err
		}
		res.RowsStrategy = strategy

		// TODO: Add support for number of rows
		size, err := parseBytes(raw.Rows.Size)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}
		res.RowsLimitBytes = size
	}

	return res, nil
}

func parseStrategy(s string) (ExtractPolicyStrategy, error) {
	switch strings.ToLower(s) {
	case "head":
		return ExtractPolicyStrategyHead, nil
	case "tail":
		return ExtractPolicyStrategyTail, nil
	default:
		return ExtractPolicyStrategyUnspecified, fmt.Errorf("invalid extract strategy %q", s)
	}
}

func parseBytes(str string) (uint64, error) {
	var s datasize.ByteSize
	if err := s.UnmarshalText([]byte(str)); err != nil {
		return 0, err
	}

	return s.Bytes(), nil
}
