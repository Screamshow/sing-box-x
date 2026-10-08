//go:build with_utls

package tls

import (
 "context"
 "crypto/ecdsa"
 "crypto/elliptic"
 "crypto/rand"
 stdtls "crypto/tls"
 "crypto/x509"
 "math/big"
 "net"
 "testing"
 "time"
 "io"

 utls "github.com/metacubex/utls"
 "github.com/stretchr/testify/require"
)

// A completed Firefox handshake must release its objects, remain usable for data,
// and accept a repeated handshake call with or without an ALPN override.
func TestFirefoxHandshakeMemoryAndReuse(t *testing.T) {
 for _, alpn := range [][]string{nil, {"h2", "http/1.1"}} {
  t.Run(string(rune('0'+len(alpn))), func(t *testing.T) {
   key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
   require.NoError(t, err)
   cert := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), DNSNames: []string{"example.test"}}
   der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
   require.NoError(t, err)
   listener, err := net.Listen("tcp", "127.0.0.1:0")
   require.NoError(t, err)
   defer listener.Close()
   served := make(chan error, 1)
   go func() {
    raw, err := listener.Accept()
    if err != nil { served <- err; return }
    defer raw.Close()
    raw.SetDeadline(time.Now().Add(10*time.Second))
    server := stdtls.Server(raw, &stdtls.Config{Certificates: []stdtls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}, NextProtos: []string{"h2", "http/1.1"}})
    message := make([]byte, 4)
    _, err = io.ReadFull(server, message)
    if err == nil { _, err = server.Write(message) }
    served <- err
   }()
   raw, err := net.Dial("tcp", listener.Addr().String())
   require.NoError(t, err)
   defer raw.Close()
   raw.SetDeadline(time.Now().Add(10*time.Second))
   conn := &utlsALPNWrapper{utlsConnWrapper{utls.UClient(raw, &utls.Config{ServerName:"example.test", InsecureSkipVerify:true}, utls.HelloFirefox_Auto)}, alpn}
   require.NoError(t, conn.HandshakeContext(context.Background()))
   require.Nil(t, conn.HandshakeState.Hello)
   require.Nil(t, conn.HandshakeState.ServerHello)
   require.Nil(t, conn.Extensions)
   require.NoError(t, conn.HandshakeContext(context.Background()))
   require.True(t, conn.ConnectionState().HandshakeComplete)
   if len(alpn)>0 { require.Equal(t,"h2",conn.ConnectionState().NegotiatedProtocol) }
   _, err = conn.Write([]byte("ping"))
   require.NoError(t, err)
   reply := make([]byte,4)
   _, err = io.ReadFull(conn,reply)
   require.NoError(t,err)
   require.Equal(t,"ping",string(reply))
   require.NoError(t,<-served)
  })
 }
}
