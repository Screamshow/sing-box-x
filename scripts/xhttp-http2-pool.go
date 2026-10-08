package v2rayxhttp

import (
    "context"
    "net"
    "net/http"
    "sync"

    "golang.org/x/net/http2"
)

// settingsConnPool publishes a connection only after a SETTINGS/PING round
// trip. Without this barrier the first packet upload can race the peer's
// header limit and turn a session-local configuration error into a GOAWAY.
// Each XMUX client owns one pool. Session cancellation aborts its dial; waiters
// can retry, and shutdown reaches pending connections through connTracker.
type settingsConnPool struct {
    access sync.Mutex
    transport *http2.Transport
    dial func(context.Context) (net.Conn, error)
    client *http2.ClientConn
    dialing chan struct{}
}

func (p *settingsConnPool) GetClientConn(request *http.Request, _ string) (*http2.ClientConn, error) {
    ctx := request.Context()
    for {
        p.access.Lock()
        if p.client != nil && p.client.ReserveNewRequest() {
            client := p.client
            p.access.Unlock()
            return client, nil
        }
        if p.dialing != nil {
            ready := p.dialing
            p.access.Unlock()
            select {
            case <-ready:
                continue
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
        ready := make(chan struct{})
        p.dialing = ready
        p.access.Unlock()
        conn, err := p.dial(ctx)
        var client *http2.ClientConn
        if err == nil {
            client, err = p.transport.NewClientConn(conn)
            if err == nil {
                err = client.Ping(ctx)
            }
            if err != nil {
                if client != nil { client.Close() }
                conn.Close()
            }
        }
        p.access.Lock()
        if err == nil { p.client = client }
        p.dialing = nil
        close(ready)
        p.access.Unlock()
        if err != nil { return nil, err }
    }
}

func (p *settingsConnPool) MarkDead(client *http2.ClientConn) {
    p.access.Lock()
    if p.client == client { p.client = nil }
    p.access.Unlock()
}
