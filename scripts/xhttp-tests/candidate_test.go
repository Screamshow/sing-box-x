//go:build with_utls

package v2rayxhttp

import (
    "context"
    "testing"

    "github.com/sagernet/sing-box/common/tls"
    "github.com/sagernet/sing-box/log"
    "github.com/sagernet/sing-box/option"
    "github.com/sagernet/sing/common/json/badoption"
    M "github.com/sagernet/sing/common/metadata"
    N "github.com/sagernet/sing/common/network"
    "github.com/stretchr/testify/require"
)

func realityConfig(t *testing.T, alpn []string) tls.Config {
    t.Helper()
    config, err := tls.NewRealityClient(context.Background(), log.StdLogger(), "example.test",
        option.OutboundTLSOptions{Enabled: true, ServerName: "example.test", ALPN: alpn,
            UTLS: &option.OutboundUTLSOptions{Enabled: true, Fingerprint: "firefox"},
            Reality: &option.OutboundRealityOptions{Enabled: true, PublicKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}})
    require.NoError(t, err)
    return config
}

func TestCandidateAutoREALITYAndDownload(t *testing.T) {
    config := realityConfig(t, nil)
    for _, download := range []bool{false, true} {
        options := option.V2RayXHTTPOptions{Mode: "auto"}
        expected := modeStreamOne
        if download {
            expected = modeStreamUp
            options.Download = &option.V2RayXHTTPDownloadOptions{ServerOptions: option.ServerOptions{Server: "127.0.0.1", ServerPort: 443}}
        }
        transport, err := NewClient(context.Background(), N.SystemDialer, M.ParseSocksaddr("127.0.0.1:443"), options, config)
        require.NoError(t, err)
        client := transport.(*Client)
        require.Equal(t, expected, client.mode)
        require.Equal(t, "2", decideHTTPVersion(config, isReality(config)))
        require.NoError(t, client.Close())
    }
}

func TestCandidateHTTP3RejectedOnEitherLeg(t *testing.T) {
    ctx := context.Background()
    config, err := tls.NewClient(ctx, log.StdLogger(), "example.test",
        option.OutboundTLSOptions{Enabled: true, ALPN: badoption.Listable[string]{"h3"}})
    require.NoError(t, err)
    for _, tlsConfig := range []tls.Config{config, realityConfig(t, []string{"h3"})} {
        _, err = NewClient(ctx, N.SystemDialer, M.ParseSocksaddr("127.0.0.1:443"), option.V2RayXHTTPOptions{}, tlsConfig)
        require.ErrorContains(t, err, "HTTP/3 is not supported")
    }
    options := option.V2RayXHTTPOptions{Download: &option.V2RayXHTTPDownloadOptions{
        ServerOptions: option.ServerOptions{Server: "127.0.0.1", ServerPort: 443},
        OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{TLS: &option.OutboundTLSOptions{Enabled: true, ALPN: badoption.Listable[string]{"h3"}}},
    }}
    _, err = NewClient(ctx, N.SystemDialer, M.ParseSocksaddr("127.0.0.1:443"), options, nil)
    require.ErrorContains(t, err, "HTTP/3 is not supported")
}
