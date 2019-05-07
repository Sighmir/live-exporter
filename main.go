package main

import (
	"os"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type job struct {
	Interval int64  `yaml:"interval"`
	Live     string `yaml:"live"`
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
}

type conf struct {
	Interval int64  `yaml:"interval"`
	Jobs     []job  `yaml:"jobs"`
	Live     string `yaml:"live"`
	Type     string `yaml:"type"`
}

type label struct {
	Key   string
	Value string
}

type metric struct {
	Src    string `json:"__type"`
	Time   string
	Date   string
	Metric string
	Labels []label
	Value  string
	Help   string
	Type   string
}

func getDir() string {
	dir, err := os.Getwd()
  if err != nil {
    log.Fatalf("getDir err    #%v", err)
	}
	return dir
}

func getConfPath() string {
	var config string
	flag.StringVar(&config, "config", getDir()+"/live_exporter.yml", "Path to the config file.")
	flag.StringVar(&config, "c", getDir()+"/live_exporter.yml", "Path to the config file.")
	flag.Parse()

	return config
}

func (c *conf) getConf() *conf {
	yamlFile, err := ioutil.ReadFile(getConfPath())
	if err != nil {
		log.Printf("getConf err    #%v", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("YAML Unmarshal: %v", err)
	}

	return c
}

func (j *job) defaults(c conf) *job {
	if j.Live == "" {
		j.Live = c.Live
	}
	if j.Interval == 0 {
		j.Interval = c.Interval
	}
	if j.Type == "" {
		j.Type = c.Type
	}

	return j
}

func getMetrics(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("getMetrics err    #%v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("IO Reader: %v", err)
	}

	return string(body)
}

func parseMetrics(metrics string, src string) string {
	var mhelp string
	var mtype string
	var parsed []metric

	dt := time.Now()
	date := dt.Format("01-02-2006")
	time := dt.Format("15:04:05")

	scanner := bufio.NewScanner(strings.NewReader(metrics))

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if fields[1] == "HELP" {
			mhelp = strings.Join(fields[3:], " ")
		} else if fields[1] == "TYPE" {
			mtype = fields[len(fields)-1]
		} else {
			var m metric
			m.Src = src
			m.Date = date
			m.Time = time
			m.Type = mtype
			m.Help = mhelp
			m.Value = fields[len(fields)-1]

			fields[0] = strings.Replace(fields[0], "{", " ", 1)
			fields[0] = strings.Replace(fields[0], "}", " ", 1)
			subfields := strings.Fields(fields[0])
			m.Metric = subfields[0]

			if len(subfields) > 1 {
				labels := strings.Split(subfields[1], "\",")
				for _, la := range labels {
					li := strings.Split(la, "=")
					var l label
					l.Key = li[0]
					l.Value = strings.Replace(li[1], "\"", "", -1)
					m.Labels = append(m.Labels, l)
				}
			} else {
				m.Labels = []label{}
			}

			parsed = append(parsed, m)
		}
	}

	json, err := json.Marshal(parsed)
	if err != nil {
		log.Fatalf("JSON Marshal: %v", err)
	}

	return string(json)
}

func postMetrics(endpoint string, data string) string {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		log.Fatalf("main err    #%v", err)
	}

	_, err = fmt.Fprintf(conn, data+"\n")
	if err != nil {
		log.Fatalf("postMetrics err    #%v", err)
	}

	return "Message sent: " + data
}

func (j *job) spawnWorker() (*time.Ticker, chan struct{}) {
	ticker := time.NewTicker(time.Duration(j.Interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				metrics := getMetrics(j.URL)
				json := parseMetrics(metrics, j.Type)
				postMetrics(j.Live, json)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return ticker, quit
}

func main() {
	var workers []chan struct{}
	var c conf
	c.getConf()

	for _, j := range c.Jobs {
		j.defaults(c)
		_, quit := j.spawnWorker()
		workers = append(workers, quit)
	}

	for _, w := range workers {
		<-w
	}
}
