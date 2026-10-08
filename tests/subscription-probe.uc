// Input files are private generated Forkop/subscription JSON, never test fixtures.
let fs = require("fs");
let generated = json(fs.readfile(ARGV[0]));
let raw = json(fs.readfile(ARGV[1]));
let directory = ARGV[2];
let portBase = int(ARGV[3] || "15000");
let nodes = [];
for (let o in generated.outbounds)
    if (o.type == "vless" || o.type == "vmess" || o.type == "trojan" || o.type == "shadowsocks")
        push(nodes, o);
// Exercise gRPC omitted by the user's flag filter without modifying that policy.
for (let o in raw.outbounds || [])
    if (o.transport?.type == "grpc")
        push(nodes, o);
let index = 0;
let manifest = "";
for (let node in nodes) {
    // These fields describe the subscription card, not a sing-box outbound.
    delete node.remark;
    delete node.__forkop_description;
    delete node.share_link;
    index++;
    let port = portBase + index;
    let config = {
        log: { level: "warn" },
        dns: {
            servers: filter(generated.dns.servers, server => server.type != "fakeip"),
            final: "dns-server",
            strategy: generated.dns.strategy
        },
        inbounds: [{ type: "mixed", tag: "probe", listen: "127.0.0.1", listen_port: port }],
        outbounds: [node],
        route: {
            final: node.tag,
            default_domain_resolver: "dns-server",
            auto_detect_interface: generated.route.auto_detect_interface,
            default_mark: generated.route.default_mark
        }
    };
    fs.writefile(directory + "/node-" + index + ".json", sprintf("%J", config));
    manifest += index + "\t" + port + "\t" + (node.transport?.type || "tcp") + "\t" + (node.tls?.utls?.fingerprint || "none") + "\n";
}
fs.writefile(directory + "/nodes.tsv", manifest);
print("probe_nodes=", index, "\n");
