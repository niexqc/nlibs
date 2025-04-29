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

func XlsxReadCell(fileName string, sheetName, cell string) (valStr string, err error) {
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return valStr, err
	}
	defer f.Close()
	val, err := f.GetCellValue(sheetName, cell)
	return val, err
}

func XlsxReadCells(fileName string, sheetName string, cells []string) (vals []string, err error) {
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return vals, err
	}
	defer f.Close()
	for _, cell := range cells {
		val, err := f.GetCellValue(sheetName, cell)
		if err != nil {
			return vals, err
		}
		vals = append(vals, val)
	}

	return vals, err
}
