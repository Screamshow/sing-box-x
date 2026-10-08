package transport

import (
 "context"
 "io"
 "net"
 "net/http"
 "net/http/httptest"
 "net/url"
 "testing"
 "time"

 "github.com/sagernet/sing-box/dns"
 M "github.com/sagernet/sing/common/metadata"
 mdns "github.com/miekg/dns"
 "github.com/stretchr/testify/require"
)

type testDNSDialer struct { net.Dialer }
func (d testDNSDialer) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn,error) {
 return d.Dialer.DialContext(ctx,network,destination.String())
}
func (testDNSDialer) ListenPacket(context.Context,M.Socksaddr) (net.PacketConn,error) { panic("not used") }

func TestDoHTimeoutPreservesLiveTransportAndResetsFrozenOne(t *testing.T) {
 slowStarted:=make(chan struct{},2)
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  raw,err:=io.ReadAll(r.Body)
  if err!=nil { return }
  query:=new(mdns.Msg)
  if query.Unpack(raw)!=nil { return }
  if query.Question[0].Name=="slow.test." {
   slowStarted<-struct{}{}
   <-r.Context().Done()
   return
  }
  response:=new(mdns.Msg)
  response.SetReply(query)
  data,_:=response.Pack()
  w.Header().Set("Content-Type",MimeType)
  w.Write(data)
 }))
 defer server.Close()
 destination,err:=url.Parse(server.URL)
 require.NoError(t,err)
 transport:=NewHTTPSRaw(dns.TransportAdapter{},nil,testDNSDialer{},destination,http.Header{},M.ParseSocksaddr(destination.Host),nil)
 defer transport.Close()
 query:=func(name string)*mdns.Msg { q:=new(mdns.Msg); q.SetQuestion(name,mdns.TypeA); return q }
 original:=transport.transport
 timeout:=func() <-chan error {
  done:=make(chan error,1)
  go func() { ctx,cancel:=context.WithTimeout(context.Background(),500*time.Millisecond); defer cancel(); _,err:=transport.Exchange(ctx,query("slow.test.")); done<-err }()
  select { case <-slowStarted: case <-time.After(3*time.Second): t.Fatal("slow query did not start") }
  return done
 }
 first:=timeout()
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
 defer cancel()
 _,err=transport.Exchange(ctx,query("fast.test."))
 require.NoError(t,err)
 require.ErrorIs(t,<-first,context.DeadlineExceeded)
 require.Same(t,original,transport.transport,"a successful later query keeps the transport alive")
 second:=timeout()
 require.ErrorIs(t,<-second,context.DeadlineExceeded)
 require.NotSame(t,original,transport.transport,"without a later answer a timeout must reset the transport")
 _,err=transport.Exchange(ctx,query("recovered.test."))
 require.NoError(t,err)
}
