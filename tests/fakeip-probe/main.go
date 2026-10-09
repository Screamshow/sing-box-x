// Scoped loopback integration probe; never reads the installed configuration.
package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/miekg/dns"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func check(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func query(name string) string {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), dns.TypeA)
	r, _, e := (&dns.Client{Timeout: 3 * time.Second}).Exchange(m, "127.0.0.1:48153")
	must(e)
	check(r.Rcode == dns.RcodeSuccess, "DNS response failure")
	for _, a := range r.Answer {
		if a, ok := a.(*dns.A); ok {
			check(strings.HasPrefix(a.A.String(), "198.18.") || strings.HasPrefix(a.A.String(), "198.19."), "outside fakeip range")
			return a.A.String()
		}
	}
	panic("missing A answer")
}
func socks(ip string, port int) (net.Conn, error) {
	c, e := net.DialTimeout("tcp", "127.0.0.1:48154", 3*time.Second)
	if e != nil {
		return nil, e
	}
	c.SetDeadline(time.Now().Add(5 * time.Second))
	_, e = c.Write([]byte{5, 1, 0})
	if e != nil {
		c.Close()
		return nil, e
	}
	b := make([]byte, 2)
	_, e = io.ReadFull(c, b)
	if e != nil {
		c.Close()
		return nil, e
	}
	check(b[1] == 0, "SOCKS auth")
	req := append([]byte{5, 1, 0, 1}, net.ParseIP(ip).To4()...)
	req = append(req, byte(port>>8), byte(port))
	_, e = c.Write(req)
	if e != nil {
		c.Close()
		return nil, e
	}
	b = make([]byte, 4)
	_, e = io.ReadFull(c, b)
	if e != nil {
		c.Close()
		return nil, e
	}
	if b[1] != 0 {
		c.Close()
		return nil, fmt.Errorf("SOCKS refusal %d", b[1])
	}
	size := 6
	if b[3] == 4 {
		size = 18
	}
	_, e = io.ReadFull(c, make([]byte, size))
	if e != nil {
		c.Close()
		return nil, e
	}
	return c, nil
}
func fetch(ip string, secure bool, cert *x509.CertPool) {
	port := 48155
	if secure {
		port = 48156
	}
	c, e := socks(ip, port)
	must(e)
	defer c.Close()
	if secure {
		t := tls.Client(c, &tls.Config{ServerName: "fakeip.test", RootCAs: cert, MinVersion: tls.VersionTLS13})
		must(t.Handshake())
		c = t
	}
	_, e = c.Write([]byte("GET / HTTP/1.1\r\nHost: fakeip.test\r\nConnection: close\r\n\r\n"))
	must(e)
	r, e := http.ReadResponse(bufio.NewReader(c), nil)
	must(e)
	b, e := io.ReadAll(r.Body)
	must(e)
	r.Body.Close()
	check(r.StatusCode == 200 && string(b) == "FAKEIP_OK", "wrong destination or body")
}
func main() {
	binary, dir := os.Args[1], os.Args[2]
	must(os.MkdirAll(dir, 0700))
	key, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	must(e)
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), DNSNames: []string{"fakeip.test"}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, e := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	must(e)
	raw, e := x509.MarshalECPrivateKey(key)
	must(e)
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	must(os.WriteFile(certPath, certPEM, 0600))
	must(os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: raw}), 0600))
	pool := x509.NewCertPool()
	check(pool.AppendCertsFromPEM(certPEM), "certificate")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check(r.Host == "fakeip.test", "wrong host")
		io.WriteString(w, "FAKEIP_OK")
	})
	httpServer := &http.Server{Addr: "127.0.0.1:48155", Handler: handler}
	tlsServer := &http.Server{Addr: "127.0.0.1:48156", Handler: handler}
	l, e := net.Listen("tcp", httpServer.Addr)
	must(e)
	lt, e := net.Listen("tcp", tlsServer.Addr)
	must(e)
	go httpServer.Serve(l)
	go tlsServer.ServeTLS(lt, certPath, keyPath)
	defer httpServer.Close()
	defer tlsServer.Close()
	packet, e := net.ListenPacket("udp", "127.0.0.1:48158")
	must(e)
	resolver := &dns.Server{PacketConn: packet, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		if request.Question[0].Name == "slow.test." {
			time.Sleep(600 * time.Millisecond)
		}
		answer := new(dns.Msg)
		answer.SetReply(request)
		answer.Answer = []dns.RR{&dns.TXT{Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1}, Txt: []string{"OK"}}}
		w.WriteMsg(answer)
	})}
	go resolver.ActivateAndServe()
	defer resolver.Shutdown()
	cfg := map[string]any{
		"log":          map[string]any{"level": "info"},
		"dns":          map[string]any{"servers": []any{map[string]any{"type": "fakeip", "tag": "fake", "inet4_range": "198.18.0.0/15"}, map[string]any{"type": "udp", "tag": "resolver", "server": "127.0.0.1", "server_port": 48158}}, "rules": []any{map[string]any{"query_type": []string{"A", "AAAA"}, "server": "fake"}}, "final": "resolver", "disable_cache": true},
		"inbounds":     []any{map[string]any{"type": "direct", "tag": "dns-in", "listen": "127.0.0.1", "listen_port": 48153, "override_address": "127.0.0.1", "override_port": 53}, map[string]any{"type": "socks", "listen": "127.0.0.1", "listen_port": 48154}},
		"outbounds":    []any{map[string]any{"type": "direct", "tag": "direct", "routing_mark": 134217728}},
		"route":        map[string]any{"rules": []any{map[string]any{"inbound": "dns-in", "action": "hijack-dns"}, map[string]any{"action": "sniff", "timeout": "500ms"}, map[string]any{"domain": "fakeip.test", "action": "route", "outbound": "direct", "override_address": "127.0.0.1"}}, "final": "direct", "default_domain_resolver": "resolver"},
		"experimental": map[string]any{"cache_file": map[string]any{"enabled": true, "path": filepath.Join(dir, "cache.db"), "store_fakeip": true}, "clash_api": map[string]any{"external_controller": "127.0.0.1:48157"}},
	}
	data, e := json.MarshalIndent(cfg, "", "  ")
	must(e)
	config := filepath.Join(dir, "config.json")
	must(os.WriteFile(config, data, 0600))
	var cmd *exec.Cmd
	var log *os.File
	stop := func(kill bool) {
		if cmd == nil {
			return
		}
		if kill {
			cmd.Process.Kill()
		} else {
			cmd.Process.Signal(os.Interrupt)
		}
		cmd.Wait()
		log.Close()
		cmd = nil
	}
	defer func() { stop(false) }()
	start := func() {
		log, e = os.OpenFile(filepath.Join(dir, "core.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		must(e)
		cmd = exec.Command(binary, "run", "-c", config)
		cmd.Stdout = log
		cmd.Stderr = log
		must(cmd.Start())
		time.Sleep(time.Second)
	}
	start()
	if os.Getenv("DNS_NAT_PROBE") == "1" {
		conn, err := dns.Dial("udp", "127.0.0.1:48153")
		must(err)
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		message := func(name string, id uint16) *dns.Msg {
			m := new(dns.Msg)
			m.SetQuestion(name, dns.TypeTXT)
			m.Id = id
			return m
		}
		must(conn.WriteMsg(message("prime.test.", 100)))
		prime, err := conn.ReadMsg()
		must(err)
		check(prime.Rcode == dns.RcodeSuccess, "prime DNS failed")
		must(conn.WriteMsg(message("slow.test.", 101)))
		time.Sleep(30 * time.Millisecond)
		begin := time.Now()
		must(conn.WriteMsg(message("fast.test.", 102)))
		fast, err := conn.ReadMsg()
		must(err)
		check(fast.Id == 102 && fast.Rcode == dns.RcodeSuccess && time.Since(begin) < 300*time.Millisecond, "slow DNS blocked fast DNS")
		slow, err := conn.ReadMsg()
		must(err)
		check(slow.Id == 101 && slow.Rcode == dns.RcodeSuccess, "slow DNS failed")
		conn.Close()
		fmt.Println("same_udp_session_slow_does_not_block_fast=PASS")
	}
	seen := map[string]bool{}
	begin := time.Now()
	for i := 0; i < 100; i++ {
		a := query(fmt.Sprintf("name%d.fakeip.test", i))
		check(!seen[a], "duplicate address")
		seen[a] = true
	}
	fmt.Printf("dns_100_unique=PASS elapsed_ms=%d\n", time.Since(begin).Milliseconds())
	ip := query("fakeip.test")
	check(query("FAKEIP.TEST") == ip, "case-insensitive lookup")
	fetch(ip, false, pool)
	fetch(ip, true, pool)
	fmt.Println("casefold_http_tls=PASS")
	stop(false)
	start()
	check(query("fakeip.test") == ip, "clean restart changed mapping")
	fmt.Println("clean_restart=PASS")
	before := query("before-crash.fakeip.test")
	stop(true)
	start()
	after := query("after-crash.fakeip.test")
	check(before != after, "crash reused allocated address")
	check(query("fakeip.test") == ip, "persisted mapping lost")
	fmt.Println("sigkill_reserve=PASS")
	request, e := http.NewRequestWithContext(context.Background(), "POST", "http://127.0.0.1:48157/cache/fakeip/flush", nil)
	must(e)
	response, e := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	must(e)
	response.Body.Close()
	check(response.StatusCode == 204 || response.StatusCode == 200, "flush failure")
	fetch(ip, false, pool)
	fetch(ip, true, pool)
	fmt.Println("flush_missing_record_http_tls=PASS")
	c, e := socks(ip, 48155)
	if e == nil {
		c.Write([]byte("SSH-2.0-fakeip-test\r\n"))
		var b [1]byte
		_, e = c.Read(b[:])
		c.Close()
		if n, ok := e.(net.Error); ok && n.Timeout() {
			panic("nameless connection timed out instead of refusal")
		}
		check(e != nil, "nameless connection accepted")
	}
	fmt.Println("missing_record_without_name=REFUSED")
	stop(false)
	logBytes, e := os.ReadFile(filepath.Join(dir, "core.log"))
	must(e)
	check(strings.Contains(string(logBytes), "no fakeip record"), "missing warning")
	fmt.Println("warning_for_missing_record=PASS")
	if os.Getenv("MEM_LIMIT_PROBE") != "" {
		check(strings.Contains(string(logBytes), "following the available memory"), "adaptive memory limit did not start")
		fmt.Println("adaptive_memory_limit=PASS")
	}
}
