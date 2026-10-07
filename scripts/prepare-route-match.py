#!/usr/bin/env python3
"""Capture inline destination matches for Clash API without changing routing."""
from pathlib import Path
import argparse


def prepare(source: Path) -> None:
    marker = source / "route/rule/rule_match_payload.go"
    if marker.exists():
        print("Route match payload already prepared")
        return
    updates = {}

    def replace(name, old, new):
        path = source / name
        text = updates.get(path, path.read_text(encoding="utf-8"))
        if text.count(old) != 1:
            raise ValueError(f"upstream changed: {name}: {old[:70]!r}")
        updates[path] = text.replace(old, new)

    replace("adapter/inbound.go", "\t// rule cache\n", "\t// Inline destination evidence captured during rule evaluation.\n\tRuleMatchPayload string\n\n\t// rule cache\n")
    replace("adapter/inbound.go", "func (c *InboundContext) ResetRuleCache() {", "func (c *InboundContext) ResetRuleCache() {\n\tc.RuleMatchPayload = \"\"")
    replace("route/rule/rule_abstract.go", "func (r *abstractDefaultRule) Match(metadata *adapter.InboundContext) bool {", "func (r *abstractDefaultRule) matchWithoutPayload(metadata *adapter.InboundContext) bool {")
    replace("route/rule/rule_abstract.go", "func (r *abstractLogicalRule) Match(metadata *adapter.InboundContext) bool {", "func (r *abstractLogicalRule) matchWithoutPayload(metadata *adapter.InboundContext) bool {\n\tvar payloads []string")
    replace("route/rule/rule_abstract.go", "\t\t\tdeferredGroups |= metadata.DeferredIPCIDRMatchGroups\n\t\t}\n\t} else {", "\t\t\tdeferredGroups |= metadata.DeferredIPCIDRMatchGroups\n\t\t\tif metadata.RuleMatchPayload != \"\" { payloads = append(payloads, metadata.RuleMatchPayload) }\n\t\t}\n\t} else {")
    replace("route/rule/rule_abstract.go", "\t\t\t\tif metadata.DeferredIPCIDRMatchGroups == 0 {\n\t\t\t\t\tdeferredGroups = 0", "\t\t\t\tif metadata.DeferredIPCIDRMatchGroups == 0 {\n\t\t\t\t\tpayloads = []string{metadata.RuleMatchPayload}\n\t\t\t\t\tdeferredGroups = 0")
    replace("route/rule/rule_abstract.go", "\tsnapshot.restore(metadata)\n\tif matched {", "\tsnapshot.restore(metadata)\n\tmetadata.RuleMatchPayload = strings.Join(payloads, \"; \")\n\tif matched {")
    replace("route/rule/rule_abstract.go", "\t\treturn it.Match(metadata)\n", "\t\tmatched := it.Match(metadata)\n\t\tif matched {\n\t\t\tif reporter, ok := it.(interface { MatchPayload(*adapter.InboundContext) string }); ok {\n\t\t\t\tif payload := reporter.MatchPayload(metadata); payload != \"\" {\n\t\t\t\t\tif metadata.RuleMatchPayload != \"\" { metadata.RuleMatchPayload += \"; \" }\n\t\t\t\t\tmetadata.RuleMatchPayload += payload\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t\treturn matched\n")
    replace("route/rule/rule_item_domain.go", "\tdescription string\n", "\tdescription string\n\tinlineDomains []string\n\tinlineSuffixes []string\n")
    replace("route/rule/rule_item_domain.go", "\t\tdescription,\n\t}, nil", "\t\tdescription,\n\t\tdomains,\n\t\tdomainSuffixes,\n\t}, nil")
    replace("route/rule/rule_item_domain.go", "\t\t\"domain/domain_suffix=<binary>\",\n", "\t\t\"domain/domain_suffix=<binary>\",\n\t\tnil, nil,\n")
    replace("route/rule/rule_item_cidr.go", "\tdescription string\n", "\tdescription string\n\tinlinePrefixes []netip.Prefix\n\tinlineValues []string\n")
    replace("route/rule/rule_item_cidr.go", "\tvar builder netipx.IPSetBuilder\n", "\tvar builder netipx.IPSetBuilder\n\tvar prefixes []netip.Prefix\n")
    replace("route/rule/rule_item_cidr.go", "\t\t\tbuilder.AddPrefix(prefix)\n", "\t\t\tbuilder.AddPrefix(prefix)\n\t\t\tprefixes = append(prefixes, prefix)\n")
    replace("route/rule/rule_item_cidr.go", "\t\t\tbuilder.Add(addr)\n", "\t\t\tbuilder.Add(addr)\n\t\t\tprefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))\n")
    replace("route/rule/rule_item_cidr.go", "\t\tdescription: description,\n\t}, nil", "\t\tdescription: description,\n\t\tinlinePrefixes: prefixes,\n\t\tinlineValues: prefixStrings,\n\t}, nil")
    # Freeze evidence in the connection snapshot. Final routes must never inherit
    # evidence from a preceding sniff/resolve/route-options action.
    replace("experimental/clashapi/connections.go", "\tvar rule string\n", "\tvar rule string\n\tvar rulePayload string\n")
    replace("experimental/clashapi/connections.go", "\tif c.Rule != nil {\n", "\tif c.Rule != nil {\n\t\trulePayload = c.Metadata.RuleMatchPayload\n")
    replace("experimental/clashapi/connections.go", '"rulePayload": "",', '"rulePayload": rulePayload,')
    templates = Path(__file__).parent / "route-match"
    updates[marker] = (templates / "rule_match_payload.go").read_text(encoding="utf-8")
    updates[source / "route/rule/rule_match_payload_test.go"] = (templates / "rule_match_payload_test.go").read_text(encoding="utf-8")
    updates[source / "experimental/clashapi/connections_payload_test.go"] = (templates / "connections_payload_test.go").read_text(encoding="utf-8")
    for path, text in updates.items():
        path.write_text(text, encoding="utf-8", newline="\n")
    print(f"Prepared route match payload: {len(updates)} files")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    prepare(parser.parse_args().source.resolve())
