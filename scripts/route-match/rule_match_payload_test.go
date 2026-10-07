package rule

import (
	"net/netip"
	"sync"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/stretchr/testify/require"
)

func payloadRule(item RuleItem) *DefaultRule {
	return &DefaultRule{abstractDefaultRule: abstractDefaultRule{
		destinationAddressItems: []RuleItem{item}, allItems: []RuleItem{item},
	}}
}

func payloadDomains(t *testing.T, domains, suffixes []string) *DefaultRule {
	t.Helper()
	item, err := NewDomainItem(domains, suffixes)
	require.NoError(t, err)
	return payloadRule(item)
}

func TestXMatchPayloadDomains(t *testing.T) {
	rule := payloadDomains(t, []string{"exact.example"}, []string{
		"dell.com", "2ip.io", "vencord.dev", "chatgpt.com", "oaistatic.com", ".subonly.example",
	})
	for _, tc := range []struct {
		host, payload string
		matched       bool
	}{
		{"chatgpt.com", "domain_suffix=chatgpt.com", true},
		{"AB.CHATGPT.COM", "domain_suffix=chatgpt.com", true},
		{"persistent.oaistatic.com", "domain_suffix=oaistatic.com", true},
		{"exact.example", "domain=exact.example", true},
		{"child.exact.example", "", false},
		{"notchatgpt.com", "", false},
		{"chatgpt.com.evil.example", "", false},
		{"subonly.example", "", false},
		{"a.subonly.example", "domain_suffix=.subonly.example", true},
	} {
		t.Run(tc.host, func(t *testing.T) {
			metadata := &adapter.InboundContext{Destination: M.Socksaddr{Fqdn: tc.host}, RuleMatchPayload: "stale"}
			require.Equal(t, tc.matched, rule.Match(metadata))
			require.Equal(t, tc.payload, metadata.RuleMatchPayload)
		})
	}
	metadata := &adapter.InboundContext{Domain: "chatgpt.com", Destination: M.Socksaddr{Fqdn: "different.example"}}
	require.True(t, rule.Match(metadata))
	require.Equal(t, "domain_suffix=chatgpt.com", metadata.RuleMatchPayload)
}

func TestXMatchPayloadKeywordAndRE2(t *testing.T) {
	regex, err := NewDomainRegexItem([]string{"^never$", `^persistent\.[a-z]+\.com$`})
	require.NoError(t, err)
	for _, tc := range []struct {
		item    RuleItem
		payload string
	}{
		{NewDomainKeywordItem([]string{"never", "oaistatic"}), "domain_keyword=oaistatic"},
		{regex, `domain_regex=^persistent\.[a-z]+\.com$`},
	} {
		metadata := &adapter.InboundContext{Domain: "PERSISTENT.OAISTATIC.COM"}
		require.True(t, payloadRule(tc.item).Match(metadata))
		require.Equal(t, tc.payload, metadata.RuleMatchPayload)
	}
}

func TestXMatchPayloadLogicalBranches(t *testing.T) {
	domain := payloadDomains(t, nil, []string{"chatgpt.com"})
	falseAfterDomain := payloadDomains(t, nil, []string{"chatgpt.com"})
	falseAfterDomain.items = []RuleItem{&fakeRuleItem{matched: false}}
	inverted := payloadDomains(t, nil, []string{"excluded.example"})
	inverted.invert = true
	for _, tc := range []struct {
		name, mode      string
		rules           []adapter.HeadlessRule
		invert, matched bool
		payload         string
	}{
		{"and source exclusion", C.LogicalTypeAnd, []adapter.HeadlessRule{domain, inverted}, false, true, "domain_suffix=chatgpt.com"},
		{"and fails after domain", C.LogicalTypeAnd, []adapter.HeadlessRule{domain, newSingleItemRule(false)}, false, false, ""},
		{"or discards failed branch", C.LogicalTypeOr, []adapter.HeadlessRule{falseAfterDomain, newSingleItemRule(true)}, false, true, ""},
		{"or second branch", C.LogicalTypeOr, []adapter.HeadlessRule{newSingleItemRule(false), domain}, false, true, "domain_suffix=chatgpt.com"},
		{"inverted positive fails", C.LogicalTypeAnd, []adapter.HeadlessRule{domain}, true, false, ""},
		{"inverted failed succeeds", C.LogicalTypeAnd, []adapter.HeadlessRule{falseAfterDomain}, true, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rule := &abstractLogicalRule{mode: tc.mode, rules: tc.rules, invert: tc.invert}
			metadata := &adapter.InboundContext{Domain: "chatgpt.com", RuleMatchPayload: "stale"}
			require.Equal(t, tc.matched, rule.Match(metadata))
			require.Equal(t, tc.payload, metadata.RuleMatchPayload)
		})
	}
	// A later rule/action cannot inherit evidence from a preceding match.
	metadata := &adapter.InboundContext{Domain: "chatgpt.com"}
	require.True(t, domain.Match(metadata))
	metadata.ResetRuleCache()
	require.Empty(t, metadata.RuleMatchPayload)
	require.True(t, newSingleItemRule(true).Match(metadata))
	require.Empty(t, metadata.RuleMatchPayload)
}

func TestXMatchPayloadCIDRResolvedDestination(t *testing.T) {
	item, err := NewIPCIDRItem(false, []string{"10.0.0.0/8", "192.0.2.0/24", "2001:db8::/32", "127.0.0.1"})
	require.NoError(t, err)
	rule := &DefaultRule{abstractDefaultRule: abstractDefaultRule{destinationIPCIDRItems: []RuleItem{item}, allItems: []RuleItem{item}}}
	for _, tc := range []struct{ ip, payload string }{
		{"192.0.2.12", "ip_cidr=192.0.2.0/24"},
		{"2001:db8::12", "ip_cidr=2001:db8::/32"},
		{"127.0.0.1", "ip_cidr=127.0.0.1"},
	} {
		metadata := &adapter.InboundContext{Destination: M.Socksaddr{Fqdn: "unrelated.example"}, DestinationAddresses: []netip.Addr{netip.MustParseAddr(tc.ip)}}
		require.True(t, rule.Match(metadata))
		require.Equal(t, tc.payload, metadata.RuleMatchPayload)
		require.False(t, metadata.Destination.Addr.IsValid())
	}
	metadata := &adapter.InboundContext{IgnoreDestinationIPCIDRMatch: true}
	require.True(t, rule.Match(metadata))
	require.NotZero(t, metadata.DeferredIPCIDRMatchGroups)
	require.Empty(t, metadata.RuleMatchPayload)
	metadata = &adapter.InboundContext{IPCIDRAcceptEmpty: true}
	require.True(t, rule.Match(metadata))
	require.Empty(t, metadata.RuleMatchPayload)
}

func TestXMatchPayloadParallelConnections(t *testing.T) {
	rule := payloadDomains(t, nil, []string{"chatgpt.com", "oaistatic.com"})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			host := "chatgpt.com"
			if i%2 == 0 {
				host = "oaistatic.com"
			}
			metadata := &adapter.InboundContext{Domain: host}
			if !rule.Match(metadata) || metadata.RuleMatchPayload != "domain_suffix="+host {
				t.Errorf("wrong per-connection payload: %q", metadata.RuleMatchPayload)
			}
		}(i)
	}
	wg.Wait()
}

func TestXMatchPayloadOpaqueAndInvertedRules(t *testing.T) {
	domain := payloadDomains(t, nil, []string{"chatgpt.com"})
	domain.invert = true
	metadata := &adapter.InboundContext{Domain: "other.example"}
	require.True(t, domain.Match(metadata))
	require.Empty(t, metadata.RuleMatchPayload)
	domain.invert = false
	domain.ruleSetItem = &RuleSetItem{setList: []adapter.RuleSet{&fakeRuleSet{matched: true}}}
	metadata = &adapter.InboundContext{Domain: "chatgpt.com"}
	require.True(t, domain.Match(metadata))
	require.Empty(t, metadata.RuleMatchPayload)

	inner := &abstractLogicalRule{mode: C.LogicalTypeOr, rules: []adapter.HeadlessRule{
		payloadDomains(t, nil, []string{"unmatched.example"}), payloadDomains(t, nil, []string{"chatgpt.com"}),
	}}
	outer := &abstractLogicalRule{mode: C.LogicalTypeAnd, rules: []adapter.HeadlessRule{inner, newSingleItemRule(true)}}
	metadata = &adapter.InboundContext{Domain: "chatgpt.com"}
	require.True(t, outer.Match(metadata))
	require.Equal(t, "domain_suffix=chatgpt.com", metadata.RuleMatchPayload)
}
