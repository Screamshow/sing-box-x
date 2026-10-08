// Clone a private generated config, preserving the production DNS and route mark.
let fs = require("fs");
let config = json(fs.readfile(ARGV[0]));
let kind = ARGV[1];
let listPath = ARGV[2];
let resultPath = ARGV[3];
let index = int(ARGV[4] || "1");
let nodes = filter(config.outbounds, o => o.type == "vless");
let selected = nodes[index - 1];
if (!selected) die("node index does not exist\n");
config.inbounds = [{ type: "mixed", tag: "list-route-probe", listen: "127.0.0.1", listen_port: 46534 }];
config.log = { level: "warn" };
config.outbounds = filter(config.outbounds, o => o.type != "urltest");
for (let outbound in config.outbounds)
    if (outbound.type == "selector") {
        outbound.outbounds = [selected.tag];
        outbound.default = selected.tag;
    }
config.experimental = {
    clash_api: { external_controller: "127.0.0.1:19090" },
    cache_file: { enabled: true, path: resultPath + ".cache", store_fakeip: true }
};
let tag = "list-route-" + kind;
if (type(config.route.rule_set) != "array") config.route.rule_set = [];
if (type(config.route.rules) != "array") config.route.rules = [];
push(config.route.rule_set, { type: "local", tag, format: kind, path: listPath });
unshift(config.route.rules, { action: "route", rule_set: [tag], outbound: selected.tag });
fs.writefile(resultPath, sprintf("%J", config));
