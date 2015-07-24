package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
        "strings"
        "regexp"
        "strconv"
        "math"
        //"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "apache2" // For Prometheus metrics.
)

var (
	listenAddress = flag.String("telemetry.address", ":9113", "Address on which to expose metrics.")
	metricsPath  = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	apache2ScrapeURI   = flag.String("apache2.scrape_uri", "http://localhost/server-status", "URI to apache2 stub status page")
	insecure         = flag.Bool("insecure", true, "Ignore server certificate if using https")
)

// Exporter collects apache2 stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI    string
	mutex  sync.RWMutex
	client *http.Client
	scrapeFailures   prometheus.Counter
	version          prometheus.Gauge
	totalRequests    prometheus.Gauge
        uptimeDays       prometheus.Gauge
        uptimeHour       prometheus.Gauge
        totalAccesses    prometheus.Gauge
        cpuUsageUser     prometheus.Gauge
        cpuUsageSystem   prometheus.Gauge
        cpuLoad          prometheus.Gauge
        totalTraffic	 *prometheus.GaugeVec
        requestBeingProcessed prometheus.Gauge
        idleWorkers      prometheus.Gauge 
        requestsPerSecond  prometheus.Gauge
        dataPerSecond    *prometheus.GaugeVec
        dataPerRequest   *prometheus.GaugeVec
        clientRequests   *prometheus.GaugeVec
        virtualHosts     *prometheus.GaugeVec
}

// NewApache2Exporter returns an initialized Exporter.
func NewApache2Exporter(uri string) *Exporter {
	return &Exporter{
		URI: uri,
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping apache2.",
		}),
		version: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "version",
			Help:      "Apache version",
		}),
                totalRequests: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "total_requests",
                        Help:      "Total no of Requests",
                }),
                uptimeDays: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "uptime_days",
                        Help:      "Apache server uptime in days",
                }),
                uptimeHour: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "uptime_hour",
                        Help:      "Apache server uptime hour, but uptime days should be countable",
                }),
                totalAccesses: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "total_accesses",
                        Help:      "Total Accesses",
                }),
                cpuUsageUser: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "cpu_usage_user",
                        Help:      "CPU Usage (User)",
                }),
                cpuUsageSystem: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "cpu_usage_system",
                        Help:      "CPU Usage (System)",
                }),
                cpuLoad: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "cpu_load",
                        Help:      "CPU Load in %",
                }),
                totalTraffic: prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "total_traffic",
                        Help:      "Total Traffic",
                },
                        []string{"bytes"},
                ),
                requestBeingProcessed: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "request_currently_being_processed",
                        Help:      "Request Currently Being Processed",
                }),
                idleWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "idle_workers",
                        Help:      "Idle Workers",
                }),
                requestsPerSecond: prometheus.NewGauge(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "requests_per_second",
                        Help:      "Requests per second",
                }),
                dataPerSecond: prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "data_per_second",
                        Help:      "Data per second",
                },
                        []string{"bytes_per_sec"},
                ),
                dataPerRequest: prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "data_per_request",
                        Help:      "Data per request",
                },      
                        []string{"bytes_per_request"},
                ), 
                clientRequests: prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "number_of_requests_from_client",
                        Help:      "Number of requests from client",
                },      
                        []string{"ip"},
                ), 
                virtualHosts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: namespace,
                        Name:      "virtual_hosts",
                        Help:      "Number of virtual hosts",
                },      
                        []string{"vhosts"},
                ),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
			},
		},
	}
}

// Describe describes all the metrics ever exported by the apache2 exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.version.Describe(ch)
	e.totalRequests.Describe(ch)
        e.uptimeDays.Describe(ch)
        e.uptimeHour.Describe(ch)
        e.totalAccesses.Describe(ch)
        e.cpuUsageUser.Describe(ch)
        e.cpuUsageSystem.Describe(ch)
        e.cpuLoad.Describe(ch)
        e.totalTraffic.Describe(ch)
        e.requestBeingProcessed.Describe(ch)
        e.idleWorkers.Describe(ch)
        e.requestsPerSecond.Describe(ch)
        e.dataPerSecond.Describe(ch)
        e.dataPerRequest.Describe(ch)
        e.clientRequests.Describe(ch)
        e.virtualHosts.Describe(ch)
	e.scrapeFailures.Describe(ch)
}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func NumberOfDuplicate(li []string, el string) int{
    j := 0
    for _, x := range li {
        if (x == el){
            j++
        }
    }
    return j
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) error {
    resp, err := e.client.Get(e.URI)
    if err != nil {
	return fmt.Errorf("Error scraping apache2: %v", err)
    }
    data, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 400 {
	if err != nil {
        	data = []byte(err.Error())
	}
	return fmt.Errorf("Status %s (%d): %s", resp.Status, resp.StatusCode, data)
    }
    //parsing result
    statusPage := string(data)
    statusPage = strings.Replace(statusPage,"\r","",-1)
    statusPage = strings.Replace(statusPage,"\n","",-1)

    reg := regexp.MustCompile(`<td><b>(.*)</b></td><td>(.*)</td><td>(.*)</td><td>_</td><td>(.*)</td><td>(.*)</td><td>(.*)</td><td>(.*)<\/td><td>(.*)</td><td>(.*)<\/td><td>(.*)<\/td><td nowrap>(.*)<\/td><td nowrap>`)
    reg_raw := regexp.MustCompile(`<dt>(.*)</dt>`)

    matches := reg.FindAllString(statusPage, -1)[0]
    matchesSlice := strings.Split(matches,"</td><td nowrap>")
    clientSlice := []string{}
    vhostSlice := []string{}

    for i, _ := range matchesSlice{
        if(math.Mod(float64(i), 2) == 0){
            clientMatchSlice := strings.Split(matchesSlice[i], "</td><td>")
            client := clientMatchSlice[len(clientMatchSlice)-1]
            if client != ""{
                clientSlice = append(clientSlice, client)
            }
         }else{
            vhostSlice = append(vhostSlice, matchesSlice[i])
         }
    }
    xs := vhostSlice
    RemoveDuplicates(&xs)
    cl := clientSlice
    RemoveDuplicates(&cl)

    matchesraw := reg_raw.FindAllString(statusPage, -1)[0]
    matchesrawSlice := strings.Split(matchesraw, "</dt>")    
    var apachever string
    var uptime_days string
    var uptime_hour string
    var totaltrafic_val string
    var totaltrafic_ext string
    var totalaccesses string
    var cpuusage_user string
    var cpuusage_system string
    var cpuload string
    var req_being_processed string
    var idle_workers string
    var requests_per_sec string
    var data_per_sec string
    var data_per_sec_ext string
    var data_per_request string
    var data_per_request_ext string

    tagreg := regexp.MustCompile(`<[^>]*>`)

    for _, value := range matchesrawSlice{
        line := tagreg.ReplaceAllString(value,"")
        if strings.Contains(line, "- "){
           tmp_ex := strings.Split(line, "- ")
           for _, val := range tmp_ex{
               line_ex_tmp := strings.Split(val, ":")
               if strings.Contains(line_ex_tmp[0], "Total accesses"){
                   totalaccesses = strings.TrimSpace(line_ex_tmp[1])
               }
               if strings.Contains(line_ex_tmp[0], "Total Traffic"){
                   totaltrafic := strings.TrimSpace(line_ex_tmp[1])
                   totaltrafic_val = strings.Split(totaltrafic, " ")[0]
                   totaltrafic_ext = strings.Split(totaltrafic, " ")[1]
               }
               if strings.Contains(line_ex_tmp[0], "CPU Usage"){
                   cpuusagestr := strings.TrimSpace(line_ex_tmp[1])
                   cpuusageUserStr := strings.Split(cpuusagestr, " ")[0]
                   cpuusage_user = cpuusageUserStr[1:len(cpuusageUserStr)]
                   cpuusageSystemStr := strings.Split(cpuusagestr, " ")[1]
                   cpuusage_system = cpuusageSystemStr[1:len(cpuusageSystemStr)]
               }
               if strings.Contains(line_ex_tmp[0], "CPU load"){
                   cpu  := strings.TrimSpace(line_ex_tmp[0])
                   cpul := strings.Split(cpu, " ")[0]
                   cpuload = cpul[:len(cpul)-1]
               }
               if strings.Contains(val, "requests/sec"){
                   reqpersec := strings.TrimSpace(val)
                   requests_per_sec = strings.Split(reqpersec, " ")[0]
               }
               if strings.Contains(val, "/second"){
                   datapersec := strings.TrimSpace(val)
                   data_per_sec = strings.Split(datapersec, " ")[0]
                   data_per_sec_ext = strings.Split(datapersec, " ")[1]
               }
               if strings.Contains(val, "/request"){
                   dataperreq := strings.TrimSpace(val)
                   data_per_request = strings.Split(dataperreq, " ")[0]
                   data_per_request_ext = strings.Split(dataperreq, " ")[1]
               }
           }
        }
        
        line_ex := strings.Split(line, ":")
        if strings.Contains(line_ex[0], "Server Version"){
            apachever = strings.TrimSpace(line_ex[1])
            apachever = strings.Split(apachever, " ")[0]
            apachever = strings.Replace(apachever, "Apache/", "", -1)
            apachever = apachever[0:3]
        }
        if strings.Contains(line_ex[0], "Server uptime"){
            uptime := strings.TrimSpace(line_ex[1])
            uptimearr := strings.Split(uptime, " ")
            uptime_days = uptimearr[0]
            uptime_hour = uptimearr[2]
        }
        if strings.Contains(line_ex[0], "requests currently being processed"){
            req_line := strings.TrimSpace(line_ex[0])
            req_line_splt := strings.Split(req_line, ",")[0]
            req_being_processed = req_line_splt[0:1]
            idle_work := strings.TrimSpace(strings.Split(req_line, ",")[1])
            idle_workers = idle_work[0:1]
        }

    } 

    apache2_version, _ := strconv.ParseFloat(apachever, 64)
    up_day, _ := strconv.ParseFloat(uptime_days, 64)
    up_hour, _ := strconv.ParseFloat(uptime_hour, 64)
    accesses, _ := strconv.ParseFloat(totalaccesses, 64)
    cpu_usage_user, _ := strconv.ParseFloat(cpuusage_user, 64)    
    cpu_usage_system, _ := strconv.ParseFloat(cpuusage_system, 64)    
    cpu_load, _ := strconv.ParseFloat(cpuload, 64)
    tot_traf_val, _ := strconv.ParseFloat(totaltrafic_val, 64)
    request_being_processed, _ := strconv.ParseFloat(req_being_processed, 64)
    idle_workr, _ := strconv.ParseFloat(idle_workers, 64)
    req_per_sec, _ := strconv.ParseFloat(requests_per_sec, 64)
    dat_per_sec, _ := strconv.ParseFloat(data_per_sec, 64)
    dat_per_req, _ := strconv.ParseFloat(data_per_request, 64)


    e.version.Set(float64(apache2_version))
    e.totalRequests.Set(float64(len(clientSlice)))
    e.uptimeDays.Set(float64(up_day))
    e.uptimeHour.Set(float64(up_hour))
    e.totalAccesses.Set(float64(accesses))
    e.cpuUsageUser.Set(float64(cpu_usage_user))
    e.cpuUsageSystem.Set(float64(cpu_usage_system))
    e.cpuLoad.Set(float64(cpu_load))
    e.totalTraffic.WithLabelValues(totaltrafic_ext).Set(float64(tot_traf_val))
    e.requestBeingProcessed.Set(float64(request_being_processed))
    e.idleWorkers.Set(float64(idle_workr))
    e.requestsPerSecond.Set(float64(req_per_sec))
    e.dataPerSecond.WithLabelValues(data_per_sec_ext).Set(float64(dat_per_sec))
    e.dataPerRequest.WithLabelValues(data_per_request_ext).Set(float64(dat_per_req))
    for _,val :=range cl{
        c := NumberOfDuplicate(clientSlice, val)
        e.clientRequests.WithLabelValues(val).Set(float64(c))
    }
    var vhoststr string
    for _, val := range xs{
        vhoststr = val+","+vhoststr
    }
    vhoststr = vhoststr[:len(vhoststr)-1]
    e.virtualHosts.WithLabelValues(vhoststr).Set(float64(len(xs)))    
    return nil
}

// Collect fetches the stats from configured nginx location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.scrape(ch); err != nil {
		log.Printf("Error scraping apache2: %s", err)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)
	}
	e.version.Collect(ch)
	e.totalRequests.Collect(ch)
        e.uptimeDays.Collect(ch)
        e.uptimeHour.Collect(ch)
        e.totalAccesses.Collect(ch)
        e.cpuUsageUser.Collect(ch)
        e.cpuUsageSystem.Collect(ch)
        e.cpuLoad.Collect(ch)
        e.totalTraffic.Collect(ch)
        e.requestBeingProcessed.Collect(ch)
        e.idleWorkers.Collect(ch)
        e.requestsPerSecond.Collect(ch)
        e.dataPerSecond.Collect(ch)
        e.dataPerRequest.Collect(ch)
        e.clientRequests.Collect(ch)
        e.virtualHosts.Collect(ch)
	return
}

func main() {
	flag.Parse()

	exporter := NewApache2Exporter(*apache2ScrapeURI)
	prometheus.MustRegister(exporter)
	http.Handle(*metricsPath, prometheus.Handler())
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    w.Write([]byte(`<html>
                <head><title>Apache2 exporter</title></head>
                <body>
                   <h1>Apache2 exporter</h1>
                   <p><a href='` + *metricsPath + `'>Metrics</a></p>
                   </body>
                </html>
              `))
	})
	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
