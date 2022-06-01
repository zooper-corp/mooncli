package display

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/client"
	"github.com/zooper-corp/mooncli/internal/tools"
	"os"
	"strings"
)

func DumpTable(data client.CollatorPool, client *client.Client, options config.TableOptions) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	fmt.Printf(
		"Chain:%v runtime:%v round: %v block:#%v\n",
		client.Chain,
		client.SpecVersion,
		client.SnapRound.Number,
		client.SnapBlock.Number,
	)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(
		table.Row{
			"Display",
			"Rank",
			"Selected",
			"Counted",
			"Blocks",
			"Blocks",
			"Balance",
			"Revokes",
			"Revokes",
			"Revokes",
		},
		rowConfigAutoMerge,
	)
	t.AppendHeader(table.Row{
		"",
		"",
		"",
		"Free",
		"Now",
		"Avg",
		"",
		"Counted",
		"Delta",
		"New Rank",
	})
	cc := []table.ColumnConfig{
		{
			Name: "Display",
			Transformer: func(val interface{}) string {
				return strings.Trim(val.(string), " ")
			},
		},
		{Name: "Rank"},
		{Name: "Selected", Hidden: true},
		{Name: "Counted"},
		{Name: "Blocks"},
		{Name: "Blocks Avg", Hidden: options.Compact},
		{Name: "Balance", Hidden: options.Compact},
		{Name: "New Counted", Hidden: options.Compact},
		{Name: "New Delta", Hidden: options.Compact},
		{Name: "New Rank", Hidden: options.Compact},
	}
	t.SetColumnConfigs(cc)
	// Add rows
	t.SetRowPainter(func(row table.Row) text.Colors {
		if !row[2].(bool) {
			return text.Colors{text.FgHiBlack}
		} else if row[1].(uint32) > data.SelectedSize {
			return text.Colors{text.FgYellow}
		}
		return nil
	})
	revokeRound := data.RoundNumber + options.RevokeRounds
	for _, info := range data.Collators {
		t.AppendRow(table.Row{
			tools.ToAscii(info.DisplayName()),
			info.Rank,
			info.Selected,
			// Counted
			tools.Humanize(info.Counted.Float64()),
			// Blocks
			fmt.Sprintf("%v", info.History[data.RoundNumber].Blocks),
			fmt.Sprintf("%.1f", info.AverageBlocks()),
			// Balance
			tools.Humanize(info.Balance.Free.Float64()),
			// Revokes
			tools.Humanize(info.RevokeAt(revokeRound).Counted.Float64()),
			tools.Humanize(info.Counted.Float64() - info.RevokeAt(revokeRound).Counted.Float64()),
			fmt.Sprintf("%v", info.Revokes[revokeRound].Rank),
		})
	}
	// Sort and render
	t.SetAllowedRowLength(options.GetTableWidth())
	sortIndex := 2
	if options.GetSortKey() != "" {
		sk := strings.ReplaceAll(options.GetSortKey(), " ", "")
		for i, c := range cc {
			k := strings.ReplaceAll(c.Name, " ", "")
			if strings.EqualFold(sk, k) {
				sortIndex = i + 1
				break
			}
		}
	}
	println(sortIndex)
	t.SortBy([]table.SortBy{{Number: sortIndex, Mode: options.GetSortMode()}})
	t.Render()
}
