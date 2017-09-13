# Apache2 Exporter for Prometheus

![image](https://user-images.githubusercontent.com/7112075/30365549-acc919e0-989a-11e7-9c31-b9b7e9a5b036.png)

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
