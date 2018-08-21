package stconfig

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

type Sheet struct {
	Row  uint16
	Col  uint16
	Data [][]string
}

func ReadXlsx(file string, sheet string) (*Sheet, error) {
	wk, err := xlsx.OpenFile(file)
	if err != nil {
		return nil, err
	}
	if len(wk.Sheets) > 0 {
		var st *xlsx.Sheet
		if sheet == "" {
			st = wk.Sheets[0]
		} else {
			st = wk.Sheet[sheet]
		}
		if st != nil {
			data := make([][]string, st.MaxRow)
			for r, row := range st.Rows {
				data[r] = make([]string, st.MaxCol)
				for c, cell := range row.Cells {
					data[r][c] = cell.String()
				}
			}
			return &Sheet{uint16(st.MaxRow), uint16(st.MaxCol), data}, nil
		}
	}

	return nil, fmt.Errorf("can not find sheet %s", sheet)
}
