package config

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

type TableOptions struct {
	Compact      bool
	SortKey      string
	SortDesc     bool
	RevokeRounds uint32
}

func GetDefaultTableOptions() TableOptions {
	return TableOptions{
		Compact:      false,
		SortKey:      "Rank",
		SortDesc:     false,
		RevokeRounds: 28,
	}
}

func (to *TableOptions) GetSortKey() string {
	return cases.Title(language.Und).String(to.SortKey)
}

func (to *TableOptions) GetTableWidth() int {
	if to.Compact {
		return 80
	} else {
		return 120
	}
}

func (to *TableOptions) GetSortMode() table.SortMode {
	if strings.EqualFold(to.SortKey, "rank") ||
		strings.EqualFold(to.SortKey, "balance") ||
		strings.EqualFold(to.SortKey, "blocks") ||
		strings.EqualFold(to.SortKey, "blocks avg") {
		if to.SortDesc {
			return table.DscNumeric
		} else {
			return table.AscNumeric
		}
	} else {
		if to.SortDesc {
			return table.Dsc
		} else {
			return table.Dsc
		}
	}
}
