package ntools

import (
	"github.com/xuri/excelize/v2"
)

func XlsxRead(fileName string, sheetName string, dataIdx int) (contents [][]string, err error) {
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return contents, err
	}
	defer f.Close()
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return contents, err
	}
	for idx, v := range rows {
		if idx < dataIdx {
			continue
		}
		contents = append(contents, v)
	}
	return contents, err
}
