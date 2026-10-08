package v2rayxhttp

import "testing"

// Repeat the donor regression on fresh connections: the first upload must see
// the peer's advertised header limit, preserving other sessions on that socket.
func TestCandidateFirstSettings(t *testing.T) {
    for i := 0; i < 20; i++ {
        TestXHTTPHeadersTooLarge(t)
    }
}
