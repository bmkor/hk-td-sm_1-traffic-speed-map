package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// TSDTimeStamps : trafficspeeddata timestamps response struct
type TSDTimeStamps struct {
	Timestamps   []string `json:"timestamps"`
	VersionCount int      `json:"version-count"`
}

var (
	// this is where you can specify how many maxFileDescriptors
	// you want to allow open
	maxFileDescriptors = 100
	// wait time (in sec) between batch download
	// Warning: too short will result in drop connection
	waittimeinsecond = 2
	// download file extension
	// default: xml
	fileExt = "xml"
)

// rurlQ : Query url string
var rurlQ string

func getTimestamps(start string, end string) *TSDTimeStamps {
	// from := "20190101"
	// to := "20190102"
	recoverFromPanic()
	startQ := url.QueryEscape(start)
	endQ := url.QueryEscape(end)

	url := url.URL{
		Scheme: "https",
		Host:   "api.data.gov.hk",
		Path:   "v1/historical-archive/list-file-versions",
	}
	q := url.Query()
	q.Set("start", startQ)
	q.Set("end", endQ)
	q.Set("url", rurlQ)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		Logger.Critical(err)
		panic(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		Logger.Critical(err)
		panic(err)
	}

	defer resp.Body.Close()
	ts := &TSDTimeStamps{}
	if err := json.NewDecoder(resp.Body).Decode(&ts); err != nil {
		Logger.Critical(err)
		panic(err)
	}
	Logger.Infof("Timestamps (count %d) for start %s and end %s successfully retrieved.", ts.VersionCount, start, end)
	return ts
}

func downloadTSDXML(ts *TSDTimeStamps, ddir string) int {

	chunkSize := ts.VersionCount / maxFileDescriptors
	if chunkSize < 1 {
		chunkSize = 1
	}
	var pts [][]string
	for i := 0; i < ts.VersionCount; i += chunkSize {
		end := i + chunkSize
		if end > ts.VersionCount {
			end = ts.VersionCount
		}
		pts = append(pts, ts.Timestamps[i:end])
	}

	// tcnt : total count of errors
	tcnt := 0
	for i, pt := range pts {
		errch := make(chan error, len(pt))
		Logger.Infof("Chunk %d (no. of files: %d) starts downloading...", i, len(pt))
		for _, t := range pt {
			go func(t string) {
				url := url.URL{
					Scheme: "https",
					Host:   "api.data.gov.hk",
					Path:   "v1/historical-archive/get-file",
				}
				q := url.Query()
				q.Set("url", rurlQ)
				q.Set("time", t)
				url.RawQuery = q.Encode()
				tt, err := time.Parse("20060102-1504", t)
				if err != nil {
					Logger.Critical(err)
				}
				dir1 := tt.Format("2006")
				dir2 := tt.Format("01")
				dir3 := tt.Format("02")

				err = writeTSDXML(url.String(), filepath.Join(ddir, dir1, dir2, dir3, t+"."+fileExt))
				if err != nil {
					Logger.Errorf("Download file %s error %s", t+".xml", err.Error())
					errch <- err
					return
				}
				errch <- nil
				Logger.Infof("Download file %s succeeded.", t+".xml")
			}(t)
		}
		cnt := 0
		for i := 0; i < len(pt); i++ {
			if err := <-errch; err != nil {
				tcnt++
				cnt++
			}
		}
		Logger.Infof("Number of errors: %d", cnt)
		time.Sleep(time.Duration(waittimeinsecond) * time.Second)
	}
	return tcnt
}

func writeTSDXML(url string, fn string) error {
	Logger.Infof("Start downloading file %s...", filepath.Base(fn))
	// resp, err := http.Get(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Logger.Errorf("Request error %s",err.Error())
		return err
	}

	client := &http.Client{}
	if err != nil {
		// Logger.Errorf("Create client error %s",err.Error())
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		// Logger.Errorf("Get response error %s",err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("invalid status quote %d", resp.StatusCode)
	}
	os.MkdirAll(filepath.Dir(fn), os.ModePerm)

	out, err := os.Create(fn)
	if err != nil {
		resp.Body.Close()
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		resp.Body.Close()
		return err
	}
	err = out.Close()
	if err != nil {
		resp.Body.Close()
		return err
	}

	resp.Body.Close()
	return nil
}
