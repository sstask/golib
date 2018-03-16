package stconfig

import (
	"fmt"

	"github.com/extrame/xls"
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

func ReadXls(file string, sheet string) (*Sheet, error) {
	wk, err := xls.Open(file, "utf-8")
	if err != nil {
		return nil, err
	}
	if sheet == "" {
		return readsheet(wk, nil, 100000), nil
	} else {
		for i := 0; i < wk.NumSheets(); i++ {
			st := wk.GetSheet(i)
			if st.Name == sheet {
				return readsheet(wk, st, 100000), nil
			}
		}
	}
	return nil, fmt.Errorf("can not find sheet %s", sheet)
}

func readsheet(w *xls.WorkBook, sheet *xls.WorkSheet, max int) (sh *Sheet) {
	if sheet == nil {
		sheet = w.GetSheet(0)
	}
	if sheet.MaxRow != 0 {
		leng := int(sheet.MaxRow) + 1
		if max < leng {
			leng = max
		}

		temp := make([][]string, leng)
		maxcollen := uint16(0)
		for k, row := range sheet.Rows {
			data := make([]string, 0)
			if len(row.Cols) > 0 {
				for _, col := range row.Cols {
					if uint16(len(data)) <= col.LastCol() {
						data = append(data, make([]string, col.LastCol()-uint16(len(data))+1)...)
					}
					str := col.String(w)
					for i := uint16(0); i < col.LastCol()-col.FirstCol()+1; i++ {
						data[col.FirstCol()+i] = str[i]
					}
				}
				if leng > int(k) {
					temp[k] = data
					collen := uint16(len(data))
					if maxcollen < collen {
						maxcollen = collen
					}
				}
			}
		}
		for i := 0; i < len(temp); i++ {
			if uint16(len(temp[i])) < maxcollen {
				temp[i] = append(temp[i], make([]string, maxcollen-uint16(len(temp[i])))...)
			}
		}
		return &Sheet{uint16(leng), uint16(maxcollen), temp}
	}
	return nil
}
