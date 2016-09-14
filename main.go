package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var (
		graphite = flag.String("graphite_host", "127.0.0.1:2003", "host and port of carbon server")
		prefix   = flag.String("prefix", "servers", "metric prefix")
		suffix   = flag.String("suffix", "puppet.cert_age", "metric prefix")
		cert_dir = flag.String("cert_dir", "/etc/puppetlabs/puppet/ssl/ca/signed", "Puppet certificate directory")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nMore information at https://github.com/jasonhancock/puppet_cert_age\n")
	}

	flag.Parse()

	mw, err := NewMetricWriter(*graphite)
	if err != nil {
		log.Fatalln(err)
	}

	err = CheckCerts(*cert_dir, *prefix, *suffix, mw)
	if err != nil {
		log.Fatalln(err)
	}

	mw.Close()
}

type MetricWriter struct {
	conn *net.TCPConn
}

func NewMetricWriter(server string) (*MetricWriter, error) {
	addr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	m := &MetricWriter{conn: conn}
	return m, nil
}

func (m *MetricWriter) Close() {
	m.conn.Close()
}

func (m *MetricWriter) WriteMetric(metric string, value int64, ts time.Time) error {
	str := fmt.Sprintf("%s %d %d\n", metric, value, ts.Unix())
	log.Print(str)
	_, err := m.conn.Write([]byte(str))
	return err
}

func CheckCerts(dir, prefix, suffix string, m *MetricWriter) error {
	now := time.Now()

	matches, err := filepath.Glob(filepath.Join(dir, "*.pem"))
	if err != nil {
		return err
	}

	for _, cfile := range matches {
		cert, err := ParseCert(cfile)
		if err != nil {
			log.Printf("Unable to parse certificate file %s: %s\n", cfile, err)
			continue
		}

		name := fmt.Sprintf(
			"%s.%s.%s",
			prefix,
			strings.Replace(cert.Subject.CommonName, ".", "_", -1),
			suffix,
		)

		delta := now.Sub(cert.NotBefore)
		m.WriteMetric(name, int64(delta.Seconds()), now)
	}

	return nil
}

func ParseCert(file string) (*x509.Certificate, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("No PEM data found in cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
