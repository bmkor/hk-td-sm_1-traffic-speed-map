package main

import (
	"fmt"
	"time"
)

var (
	// batch Number
	// default = 5
	// Warning: Too large may result in connection drop
	batchNumber = 5
)

func partitionInputTimes(startyear string, endyear string) ([][]string, error) {
	if batchNumber < 1 {
		return [][]string{}, fmt.Errorf("invalid batch number. Expected more than 1 but found %d", batchNumber)
	}
	start, err := time.Parse("2006", startyear)
	if err != nil {
		return [][]string{}, err
	}
	end, err := time.Parse("2006", endyear)
	if err != nil {
		return [][]string{}, err
	}
	end = end.AddDate(0, 11, 31).Add(-24 * time.Hour)
	td := int(end.Sub(start).Hours() / 24)
	var d [][]string
	for i := 0; i < td; i += batchNumber {
		s := start.AddDate(0, 0, i)
		e := start.AddDate(0, 0, i+batchNumber-1)
		if e.After(end) {
			e = end
		}
		sf := s.Format("20060102")
		ef := e.Format("20060102")
		d = append(d, []string{sf, ef})
	}
	return d, nil
}
