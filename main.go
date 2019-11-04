package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	pbnotify "github.com/bmkor/gopushbullet"

	"github.com/spf13/viper"
	"github.com/yieldr/go-log/log"
)

// Logger logging all
var Logger = log.New()

// Configuration path
var (
	configpath     = "config"
	configfilename = "appconfig.yml"
)

// Notify pb notify
var (
	notify, _ = pbnotify.New(filepath.Join(configpath, configfilename))
)

var pbtitle string

func recoverFromPanic() {
	if r := recover(); r != nil {
		Logger.Criticalf("recovered from ", r)
	}
}

// AppConfig : struct for reading appconfig.yml
type AppConfig struct {
	Startyear           int    `yaml:"startyear"`
	Endyear             int    `yaml:"endyear"`
	FileExt             string `yaml:"fileExt"`
	Downloaddestination string `yaml:"downloaddestination"`
	Qurl                string `yaml:"qurl"`
	MaxFileDescriptors  int    `yaml:"maxFileDescriptors"`
	TitleForPushBullet  string `yaml:"titleForPushBullet"`
	Waittimeinsecond    int    `yaml:"waittimeinsecond"`
	BatchNumber         int    `yaml:"batchNumber"`
}

func readAppconfig() (*AppConfig, error) {
	v := viper.New()
	v.AddConfigPath(configpath)
	v.SetConfigName(strings.TrimSuffix(configfilename, filepath.Ext(configfilename)))
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var cf = &AppConfig{}
	err = v.UnmarshalKey("downloadconfig", cf)
	if err != nil {
		return nil, err
	}
	return cf, nil
}

func main() {

	sink, err := createLogger()
	if err != nil {
		notify.Notify("Create Logger ERROR", err.Error())
		panic(err)
	}
	cf, err := readAppconfig()
	if err != nil {
		notify.Notify("Read configuration ERROR", err.Error())
		panic(err)
	}
	rurlQ, err = url.QueryUnescape(cf.Qurl)
	if err != nil {
		notify.Notify(pbtitle+" ERROR", err.Error())
		panic(err)
	}
	fileExt = cf.FileExt
	maxFileDescriptors = cf.MaxFileDescriptors
	pbtitle = cf.TitleForPushBullet
	waittimeinsecond = cf.Waittimeinsecond
	batchNumber = cf.BatchNumber
	if cf.Startyear > cf.Endyear {
		err = fmt.Errorf("end year, %d, before start year, %d", cf.Endyear, cf.Startyear)
		notify.Notify(pbtitle+" ERROR", err)
		panic(err)
	}
	startyear := strconv.Itoa(cf.Startyear)
	endyear := strconv.Itoa(cf.Endyear)
	tt, err := partitionInputTimes(startyear, endyear)
	if err != nil {
		notify.Notify(pbtitle+" ERROR", err)
		panic(err)
	}
	for _, ts := range tt {
		if len(ts) < 2 {
			err = fmt.Errorf("should be two dates, found one. %s", strings.Join(ts, "_"))
			notify.Notify(pbtitle+" ERROR", err)
			Logger.Error(err)
			continue
		}
		start := ts[0]
		end := ts[1]
		n := fmt.Sprintf("Start download data from %s to %s...", start, end)
		Logger.Info(n)
		notify.Notify(pbtitle, n)
		ts := getTimestamps(start, end)
		tcnt := downloadTSDXML(ts, cf.Downloaddestination)
		n = fmt.Sprintf("Download data from %s to %s done. Total no. of errors: %d.", start, end, tcnt)
		Logger.Info(n)
		notify.Notify(pbtitle, n)
	}
	sink.Close()
}
