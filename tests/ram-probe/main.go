// Standalone integration measurement helper. No subscription credentials are fixtures.
package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const uuid = "00000000-0000-4000-8000-000000000001"

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func writeJSON(path string, v any) {
	b, e := json.MarshalIndent(v, "", "  ")
	must(e)
	must(os.WriteFile(path, b, 0600))
}
func prepare(dir string) {
	key, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	must(e)
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), DNSNames: []string{"ram.test"}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, e := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	must(e)
	keyDER, e := x509.MarshalECPrivateKey(key)
	must(e)
	must(os.WriteFile(filepath.Join(dir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600))
	must(os.WriteFile(filepath.Join(dir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600))
	writeJSON(filepath.Join(dir, "server.json"), map[string]any{
		"log":       map[string]any{"level": "warn"},
		"inbounds":  []any{map[string]any{"type": "vless", "listen": "127.0.0.1", "listen_port": 48101, "users": []any{map[string]any{"uuid": uuid}}, "tls": map[string]any{"enabled": true, "certificate_path": filepath.Join(dir, "cert.pem"), "key_path": filepath.Join(dir, "key.pem")}}},
		"outbounds": []any{map[string]any{"type": "direct"}},
	})
	writeClient(filepath.Join(dir, "local.json"), map[string]any{"type": "vless", "tag": "test", "server": "127.0.0.1", "server_port": 48101, "uuid": uuid, "tls": map[string]any{"enabled": true, "server_name": "ram.test", "insecure": true, "utls": map[string]any{"enabled": true, "fingerprint": "firefox"}}}, false)
}
func writeClient(path string, node map[string]any, real bool) {
	c := map[string]any{"log": map[string]any{"level": "warn"}, "inbounds": []any{map[string]any{"type": "mixed", "listen": "127.0.0.1", "listen_port": 48102}}, "outbounds": []any{node}, "route": map[string]any{"final": node["tag"]}, "experimental": map[string]any{"clash_api": map[string]any{"external_controller": "127.0.0.1:48103"}}}
	c["experimental"].(map[string]any)["debug"] = map[string]any{"listen": "127.0.0.1:48104"}
	if real {
		c["dns"] = map[string]any{"servers": []any{map[string]any{"type": "udp", "tag": "dns-server", "server": "77.88.8.8"}}, "final": "dns-server", "strategy": "prefer_ipv4"}
		c["route"] = map[string]any{"final": node["tag"], "default_domain_resolver": "dns-server", "auto_detect_interface": true, "default_mark": 134217728}
	}
	writeJSON(path, c)
}
func realConfig(input, output string) {
	b, e := os.ReadFile(input)
	must(e)
	type config struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	var configs []config
	if len(b) > 0 && b[0] == '[' {
		must(json.Unmarshal(b, &configs))
	} else {
		var c config
		must(json.Unmarshal(b, &c))
		configs = []config{c}
	}
	for _, c := range configs {
		for _, node := range c.Outbounds {
			t, _ := node["tls"].(map[string]any)
			r, _ := t["reality"].(map[string]any)
			u, _ := t["utls"].(map[string]any)
			if node["type"] != "vless" || r["enabled"] != true || u["fingerprint"] != "firefox" || node["transport"] != nil {
				continue
			}
			delete(node, "remark")
			delete(node, "__forkop_description")
			delete(node, "share_link")
			for key := range node {
				if strings.HasPrefix(key, "__forkop_") {
					delete(node, key)
				}
			}
			node["tag"] = "test"
			writeClient(output, node, true)
			fmt.Println("selected_tcp_reality_firefox")
			return
		}
	}
	panic("No TCP REALITY Firefox node")
}
func echoServer() {
	ln, e := net.Listen("tcp", "127.0.0.1:48100")
	must(e)
	for {
		c, e := ln.Accept()
		must(e)
		go func() {
			defer c.Close()
			b := make([]byte, 64)
			for {
				if _, e := io.ReadFull(c, b); e != nil {
					return
				}
				if _, e := c.Write(b); e != nil {
					return
				}
			}
		}()
	}
}
func socks(target string) (net.Conn, error) {
	c, e := net.DialTimeout("tcp", "127.0.0.1:48102", 10*time.Second)
	if e != nil {
		return nil, e
	}
	fail := func(e error) (net.Conn, error) { c.Close(); return nil, e }
	c.SetDeadline(time.Now().Add(12 * time.Second))
	_, e = c.Write([]byte{5, 1, 0})
	if e != nil {
		return fail(e)
	}
	b := make([]byte, 2)
	if _, e = io.ReadFull(c, b); e != nil {
		return fail(e)
	}
	if b[1] != 0 {
		return fail(fmt.Errorf("SOCKS authentication"))
	}
	host, port, e := net.SplitHostPort(target)
	if e != nil {
		return fail(e)
	}
	p, e := strconv.Atoi(port)
	if e != nil {
		return fail(e)
	}
	request := append([]byte{5, 1, 0, 3, byte(len(host))}, []byte(host)...)
	request = append(request, byte(p>>8), byte(p))
	if _, e = c.Write(request); e != nil {
		return fail(e)
	}
	b = make([]byte, 4)
	if _, e = io.ReadFull(c, b); e != nil {
		return fail(e)
	}
	if b[1] != 0 {
		return fail(fmt.Errorf("SOCKS reply %d", b[1]))
	}
	n := 0
	switch b[3] {
	case 1:
		n = 4
	case 4:
		n = 16
	case 3:
		var v [1]byte
		if _, e = io.ReadFull(c, v[:]); e != nil {
			return fail(e)
		}
		n = int(v[0])
	default:
		return fail(fmt.Errorf("SOCKS address"))
	}
	if _, e = io.CopyN(io.Discard, c, int64(n+2)); e != nil {
		return fail(e)
	}
	c.SetDeadline(time.Time{})
	return c, nil
}

type load struct {
	conns          []net.Conn
	mu             sync.Mutex
	active, errors atomic.Int64
	wg             sync.WaitGroup
	closing        atomic.Bool
}

func (l *load) add(n int, real bool) {
	var opening sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := 0; i < n; i++ {
		opening.Add(1)
		sem <- struct{}{}
		go func() {
			defer opening.Done()
			defer func() { <-sem }()
			target := "127.0.0.1:48100"
			if real {
				target = "www.gstatic.com:443"
			}
			c, e := socks(target)
			if e != nil {
				l.errors.Add(1)
				return
			}
			if real {
				tc := tls.Client(c, &tls.Config{ServerName: "www.gstatic.com", NextProtos: []string{"http/1.1"}})
				tc.SetDeadline(time.Now().Add(15 * time.Second))
				if e = tc.Handshake(); e != nil {
					c.Close()
					l.errors.Add(1)
					return
				}
				c = tc
			}
			l.mu.Lock()
			l.conns = append(l.conns, c)
			l.mu.Unlock()
			l.active.Add(1)
			l.wg.Add(1)
			go func() {
				defer l.wg.Done()
				defer l.active.Add(-1)
				defer c.Close()
				reader := bufio.NewReader(c)
				b := make([]byte, 64)
				for {
					c.SetDeadline(time.Now().Add(15 * time.Second))
					if real {
						_, e = c.Write([]byte("GET /generate_204 HTTP/1.1\r\nHost: www.gstatic.com\r\nConnection: keep-alive\r\n\r\n"))
						if e == nil {
							var r *http.Response
							r, e = http.ReadResponse(reader, nil)
							if e == nil {
								_, e = io.Copy(io.Discard, r.Body)
								r.Body.Close()
								if r.StatusCode != 204 {
									e = fmt.Errorf("HTTP %d", r.StatusCode)
								}
							}
						}
					} else {
						_, e = c.Write(b)
						if e == nil {
							_, e = io.ReadFull(c, b)
						}
					}
					if e != nil {
						if !l.closing.Load() {
							l.errors.Add(1)
						}
						return
					}
					if l.closing.Load() {
						return
					}
					delay := time.Second
					if real {
						delay = 5 * time.Second
					}
					time.Sleep(delay)
				}
			}()
		}()
	}
	opening.Wait()
}
func (l *load) close() {
	l.closing.Store(true)
	l.mu.Lock()
	for _, c := range l.conns {
		c.Close()
	}
	l.mu.Unlock()
	l.wg.Wait()
}
func status(pid int) map[string]int64 {
	b, e := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	must(e)
	m := map[string]int64{}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) > 1 {
			v, _ := strconv.ParseInt(f[1], 10, 64)
			m[strings.TrimSuffix(f[0], ":")] = v
		}
	}
	return m
}
func run(binary, config, output, label string, real, collect bool) {
	log, e := os.Create(output + ".log")
	must(e)
	defer log.Close()
	cmd := exec.Command(binary, "run", "-c", config)
	cmd.Stdout = log
	cmd.Stderr = log
	must(cmd.Start())
	defer func() { cmd.Process.Signal(os.Interrupt); cmd.Wait() }()
	client := &http.Client{Timeout: 3 * time.Second}
	ready := false
	for i := 0; i < 40; i++ {
		r, e := client.Get("http://127.0.0.1:48103/connections")
		if e == nil {
			r.Body.Close()
			ready = true
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !ready {
		panic("Core API unavailable")
	}
	file, e := os.Create(output)
	must(e)
	defer file.Close()
	fmt.Fprintln(file, "label\tphase\ttarget\tactive\tapi_connections\terrors\trss_kib\tanon_kib\tshmem_kib\tgo_memory_bytes\tfds")
	l := &load{}
	defer l.close()
	sample := func(phase string, n int) {
		r, e := client.Get("http://127.0.0.1:48103/connections")
		must(e)
		var api struct {
			Memory      int64             `json:"memory"`
			Connections []json.RawMessage `json:"connections"`
		}
		must(json.NewDecoder(r.Body).Decode(&api))
		r.Body.Close()
		m := status(cmd.Process.Pid)
		fds, e := os.ReadDir(fmt.Sprintf("/proc/%d/fd", cmd.Process.Pid))
		must(e)
		fmt.Fprintf(file, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", label, phase, n, l.active.Load(), len(api.Connections), l.errors.Load(), m["VmRSS"], m["RssAnon"], m["RssShmem"], api.Memory, len(fds))
	}
	phase := func(name string, n int) {
		time.Sleep(5 * time.Second)
		for i := 0; i < 10; i++ {
			sample(name, n)
			time.Sleep(time.Second)
		}
		fmt.Printf("%s phase=%s active=%d errors=%d\n", label, name, l.active.Load(), l.errors.Load())
		if collect {
			r, e := client.Get("http://127.0.0.1:48104/debug/gc")
			must(e)
			r.Body.Close()
			if r.StatusCode != 204 {
				panic("GC endpoint failed")
			}
			time.Sleep(2 * time.Second)
			for i := 0; i < 3; i++ {
				sample(name+"_gc", n)
				time.Sleep(time.Second)
			}
		}
	}
	phase("idle", 0)
	counts := []int{32, 128, 256}
	if real {
		counts = []int{32, 64}
	}
	previous := 0
	for _, n := range counts {
		l.add(n-previous, real)
		previous = n
		phase("held", n)
	}
	l.close()
	time.Sleep(20 * time.Second)
	phase("closed", 0)
	if l.errors.Load() > 0 {
		fmt.Printf("%s load_errors=%d\n", label, l.errors.Load())
	}
}
func main() {
	switch os.Args[1] {
	case "prepare":
		prepare(os.Args[2])
	case "echo":
		echoServer()
	case "real-config":
		realConfig(os.Args[2], os.Args[3])
	case "run":
		run(os.Args[2], os.Args[3], os.Args[4], os.Args[5], len(os.Args) > 6 && os.Args[6] == "real", len(os.Args) > 7 && os.Args[7] == "gc")
	default:
		panic("Unknown mode")
	}
}
