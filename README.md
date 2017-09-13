# Apache Exporter for Prometheus

![image](https://user-images.githubusercontent.com/7112075/30365549-acc919e0-989a-11e7-9c31-b9b7e9a5b036.png)

This is a simple server that periodically scrapes apache stats and exports them via HTTP for Prometheus
consumption.

To install it:

```bash
sudo apt-get install mercurial
git clone https://github.com/dhasthagheer/apache2_exporter.git
cd apache_exporter
make
```

To run it:

```bash
./apache_exporter [flags]
```

## Usage

```bash
$ ./apache2_exporter -h

Usage of ./apache2_exporter:
  -apache2.scrape_uri string
    	URI to apache2 stub status page (default "http://localhost/server-status")
  -insecure
    	Ignore server certificate if using https (default true)
  -log.level value
    	Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal, panic].
  -telemetry.address string
    	Address on which to expose metrics. (default ":9113")
  -telemetry.endpoint string
    	Path under which to expose metrics. (default "/metrics")
```

## Metrics 

| Metric	        | Type  | Descriptions  |
|:------------------|:------|:--------------|
| `apache_cpu_load` | gauge | CPU Load in % |
| `apache_cpu_usage_system` | gauge | CPU Usage (System) |
| `apache_cpu_usage_user` | gauge | CPU Usage (User) |
| `apache_data_per_request` | gauge | Data per request |
| `apache_data_per_second` | gauge | Data per second |
| `apache_idle_workers` | gauge | Idle Workers |
| `apache_number_of_requests_from_client` | gauge | Number of requests from client |
| `apache_request_currently_being_processed` | gauge | Request Currently Being Processed |
| `apache_requests_per_second` | gauge | Requests per second |
| `apache_total_accesses` | gauge | Total Accesses |
| `apache_total_requests` | gauge | Total no of Requests |
| `apache_total_traffic` | gauge | Total Traffic |
| `apache_uptime_days` | gauge | Apache server uptime in days |
| `apache_uptime_hours` | gauge | Apache server uptime hour, but uptime days should be countable |
| `apache_uptime_minutes` | gauge | Apache server uptime minutes, but uptime days should be countable |
| `apache_uptime_seconds` | gauge | Apache server uptime seconds, but uptime days should be countable |
| `apache_version` | gauge | Apache version |
| `apache_virtual_hosts` | gauge | Number of virtual hosts |

