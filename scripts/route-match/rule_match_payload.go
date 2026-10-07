package rule

import (
	"net/netip"
	"strings"

	"github.com/sagernet/sing-box/adapter"
)

// Evidence is local to one evaluation and one connection. Failed, inverted,
// deferred and opaque rule-set branches must not publish positive evidence.
func (r *abstractDefaultRule) Match(metadata *adapter.InboundContext) bool {
	metadata.RuleMatchPayload = ""
	matched := r.matchWithoutPayload(metadata)
	if !matched || r.invert || r.ruleSetItem != nil || metadata.DeferredIPCIDRMatchGroups != 0 {
		metadata.RuleMatchPayload = ""
	}
	return matched
}

func (r *abstractLogicalRule) Match(metadata *adapter.InboundContext) bool {
	metadata.RuleMatchPayload = ""
	matched := r.matchWithoutPayload(metadata)
	if !matched || r.invert || metadata.DeferredIPCIDRMatchGroups != 0 {
		metadata.RuleMatchPayload = ""
	}
	return matched
}

func matchDomainHost(metadata *adapter.InboundContext) string {
	if metadata.Domain != "" {
		return strings.ToLower(metadata.Domain)
	}
	return strings.ToLower(metadata.Destination.Fqdn)
}

func (r *DomainItem) MatchPayload(metadata *adapter.InboundContext) string {
	host := matchDomainHost(metadata)
	for _, value := range r.inlineDomains {
		if host == value {
			return "domain=" + value
		}
	}
	for _, value := range r.inlineSuffixes {
		if strings.HasPrefix(value, ".") {
			if strings.HasSuffix(host, value) {
				return "domain_suffix=" + value
			}
		} else if host == value || strings.HasSuffix(host, "."+value) {
			return "domain_suffix=" + value
		}
	}
	return ""
}

func (r *DomainKeywordItem) MatchPayload(metadata *adapter.InboundContext) string {
	host := matchDomainHost(metadata)
	for _, value := range r.keywords {
		if strings.Contains(host, value) {
			return "domain_keyword=" + value
		}
	}
	return ""
}

func (r *DomainRegexItem) MatchPayload(metadata *adapter.InboundContext) string {
	host := matchDomainHost(metadata)
	for _, matcher := range r.matchers {
		if matcher.MatchString(host) {
			return "domain_regex=" + matcher.String()
		}
	}
	return ""
}

func (r *IPCIDRItem) MatchPayload(metadata *adapter.InboundContext) string {
	// Only report destination CIDRs; source exclusions are not a destination
	// routing reason. Binary rule sets deliberately remain opaque.
	if r.isSource || metadata.IPCIDRMatchSource || len(r.inlinePrefixes) == 0 {
		return ""
	}
	var addresses []netip.Addr
	if metadata.DestinationAddressMatchFromResponse {
		addresses = metadata.DNSResponseAddressesForMatch()
	} else if metadata.Destination.IsIP() {
		addresses = []netip.Addr{metadata.Destination.Addr}
	} else {
		addresses = metadata.DestinationAddresses
	}
	for i, prefix := range r.inlinePrefixes {
		for _, address := range addresses {
			if prefix.Contains(address) {
				return "ip_cidr=" + r.inlineValues[i]
			}
		}
	}
	return ""
}
