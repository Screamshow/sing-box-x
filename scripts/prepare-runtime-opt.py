#!/usr/bin/env python3
"""Apply the pinned UDP queue, adaptive Go limit and rule-set update series."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path


def adapt_rule_updates(source: Path, patch: Path) -> None:
    """Adapt 0045 without importing podkop's start-empty/recovery-hook policy."""
    text = patch.read_text()
    block = text[text.index("+// updateOnce updates"):text.index(" // A rule-set download")]
    replacement = "\n".join(line[1:] for line in block.splitlines()
                            if line.startswith(("+", " "))) + "\n\n"
    replacement = replacement.replace("s.dropRules()", "s.rules = nil")
    path = source / "route/rule/rule_set_remote.go"
    original = path.read_text()
    start = original.index("func (s *RemoteRuleSet) updateOnce() {")
    end = original.index("func (s *RemoteRuleSet) fetch(", start)
    updated = original[:start] + replacement + original[end:]
    fields = "\trefs           atomic.Int32\n"
    if updated.count(fields) != 1:
        raise ValueError("RemoteRuleSet fields changed")
    updated = updated.replace(fields, fields + "\tfetchLoaded bool\n\tfailures int\n\tfailureLogged time.Time\n")
    loaded = '\teTagHeader := response.Header.Get("Etag")'
    if updated.count(loaded) != 1:
        raise ValueError("RemoteRuleSet fetch changed")
    updated = updated.replace(loaded, "\ts.fetchLoaded = true\n" + loaded)
    path.write_text(updated, newline="\n")
    path = source / "route/rule/rule_set_updater.go"
    updated = path.read_text()
    call = "\t\t\truleSet.updateOnce()"
    if updated.count(call) != 1 or updated.count("\t\t\tupdated = true\n") != 1:
        raise ValueError("RuleSetUpdater loop changed")
    updated = updated.replace(call, "\t\t\tif ruleSet.updateOnce() { updated = true }")
    updated = updated.replace("\t\t\tupdated = true\n", "")
    path.write_text(updated, newline="\n")
    args = ["git", "apply", "--include=route/rule/rule_set_update_log_test.go"]
    subprocess.run(args + ["--check", str(patch)], cwd=source, check=True)
    subprocess.run(args + [str(patch)], cwd=source, check=True)
    test_path = source / "route/rule/rule_set_update_log_test.go"
    test = test_path.read_text().replace("newTestRuleSet(server.URL)",
        '&RemoteRuleSet{tag: "test", url: server.URL, httpClient: &http.Client{}}')
    test_path.write_text(test, newline="\n")
    subprocess.run(["gofmt", "-w", "route/rule/rule_set_remote.go",
                    "route/rule/rule_set_updater.go", "route/rule/rule_set_update_log_test.go"], cwd=source, check=True)


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / "patches/runtime-opt"
    manifest = json.loads((patches / "manifest.json").read_text())
    manifest["adapter_sha256"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    marker = source / ".forkop-runtime-opt.json"
    if marker.exists():
        if json.loads(marker.read_text()) != manifest:
            raise ValueError("Runtime optimization preparation changed; use a clean source checkout")
        print("Runtime optimization series already prepared")
        return
    for entry in manifest["patches"]:
        patch = patches / entry["file"]
        if hashlib.sha256(patch.read_bytes()).hexdigest() != entry["sha256"]:
            raise ValueError(f"Runtime optimization patch checksum mismatch: {patch.name}")
        if patch.name.startswith("0045-"):
            adapt_rule_updates(source, patch)
            continue
        subprocess.run(["git", "apply", "--check", str(patch)], cwd=source, check=True)
        subprocess.run(["git", "apply", str(patch)], cwd=source, check=True)
    marker.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Prepared {len(manifest['patches'])} Runtime optimization patches")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    prepare(parser.parse_args().source.resolve())
