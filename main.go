package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/log"
)

var (
	addr    = flag.String("web.listen-address", ":8080", "The address to listen on for HTTP requests.")
	cluster = flag.String("ssl.cluster", "c1", "kubernetes cluster")
	prefix  = flag.String("ssl.prefix", "/etc/kubernetes/pki", "certs prefix dir")
)

// SslCollector struct
type SslCollector struct {
	CertsExpiredDaysLeft *prometheus.Desc
}

// Cert struct
type Cert struct {
	Name string
	Path string
}

func parsePemFile(path string) (et time.Time, err error) {
	certPEMBlock, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	//获取证书信息 -----BEGIN CERTIFICATE-----   -----END CERTIFICATE-----
	//这里返回的第二个值是证书中剩余的 block, 一般是rsa私钥 也就是 -----BEGIN RSA PRIVATE KEY 部分
	//一般证书的有效期，组织信息等都在第一个部分里
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil {
		return
	}
	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return
	}
	return x509Cert.NotAfter, nil
	//log.Printf("certFile=%s, validation time %s ~ %s", path,
	//x509Cert.NotBefore.Format("2006-01-02 15:04"), x509Cert.NotAfter.Format("2006-01-02 15:04"))
}

func getLeftDays(now, et time.Time) int64 {
	nowUnix := now.Unix()
	etUnix := et.Unix()
	return (etUnix - nowUnix) / 3600 / 24
}

// NewSslCollector func
func NewSslCollector() *SslCollector {
	return &SslCollector{
		CertsExpiredDaysLeft: prometheus.NewDesc(
			"certs_expired_days_left",
			"CertsExpiredDaysLeft",
			[]string{"cert", "cluster"},
			nil,
		)}
}

var (
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": "v0.1.0",
		},
	})
)

// Collect func
func (collector *SslCollector) Collect(ch chan<- prometheus.Metric) {
	// all certs
	certs := []Cert{
		{
			Name: "apiserver-etcd-client.crt",
			Path: fmt.Sprintf("%s/apiserver-etcd-client.crt", *prefix),
		},
		{
			Name: "apiserver-kubelet-client.crt",
			Path: fmt.Sprintf("%s/apiserver-kubelet-client.crt", *prefix),
		},
		{
			Name: "apiserver.crt",
			Path: fmt.Sprintf("%s/apiserver.crt", *prefix),
		},
		{
			Name: "ca.crt",
			Path: fmt.Sprintf("%s/ca.crt", *prefix),
		},
		{
			Name: "front-proxy-ca.crt",
			Path: fmt.Sprintf("%s/front-proxy-ca.crt", *prefix),
		},
		{
			Name: "front-proxy-client.crt",
			Path: fmt.Sprintf("%s/front-proxy-client.crt", *prefix),
		},
		{
			Name: "etcd-ca.crt",
			Path: fmt.Sprintf("%s/etcd/ca.crt", *prefix),
		},
		{
			Name: "etcd-healthcheck-client.crt",
			Path: fmt.Sprintf("%s/etcd/healthcheck-client.crt", *prefix),
		},
		{
			Name: "etcd-peer.crt",
			Path: fmt.Sprintf("%s/etcd/peer.crt", *prefix),
		},
		{
			Name: "etcd-server.crt",
			Path: fmt.Sprintf("%s/etcd/server.crt", *prefix),
		},
	}
	var now = time.Now()
	var leftDays int64
	for _, cert := range certs {
		et, err := parsePemFile(cert.Path)
		if err != nil {
			leftDays = 0
		}
		leftDays = getLeftDays(now, et)
		if leftDays < 0 {
			continue
		}
		//fmt.Println(leftDays)
		ch <- prometheus.MustNewConstMetric(collector.CertsExpiredDaysLeft, prometheus.GaugeValue, float64(leftDays), cert.Name, *cluster)
	}

}

// Describe func
func (collector *SslCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.CertsExpiredDaysLeft
}

func main() {
	flag.Parse()
	ssl := NewSslCollector()
	prometheus.MustRegister(ssl)
	prometheus.MustRegister(version)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting Prometheus Hadoop exporter")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Hadoop Exporter</title></head>
			<body>
			<h1>Hadoop Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
