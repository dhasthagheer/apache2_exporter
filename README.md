# Apache2 Exporter for Prometheus

This is a simple server that periodically scrapes apache stats and exports them via HTTP for Prometheus
consumption.

To install it:

```bash
sudo apt-get install mercurial
git clone https://github.com/dhasthagheer/apache2_exporter.git
cd apache2_exporter
make
```

To run it:

```bash
./apache2_exporter [flags]
```

Help on flags:
```bash
./apache2_exporter --help
```
