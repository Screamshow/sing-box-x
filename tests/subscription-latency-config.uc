// Isolated VLESS subscription fixture; preserves normal Forkop parsing/generation.
let fs=require("fs"); let w=ARGV[0]; let lib=w+"/repo/forkop/files/usr/lib";
let raw=json(fs.readfile(w+"/subscription.json")),base=json(fs.readfile(w+"/base.json"));
function q(s) { return "'"+replace(""+s,/'/g,"'\\''")+"'"; }
let fixture={settings:{".name":"settings",".type":"settings",dns_server:"1.1.1.1"},section:[{".name":"probe",".type":"section",enabled:"1",action:"connection",outbound_jsons:map(raw.outbounds,o=>sprintf("%J",o))}]};
fs.writefile(w+"/fixture.json",sprintf("%J",fixture));
assert(system(join(" ",map(["env","FORKOP_LIB="+lib,"FORKOP_RUNTIME_STATE_DIR="+w+"/runtime", "ucode","-L",lib,lib+"/singbox/generator.uc","generate-config-fixture",w+"/fixture.json",w+"/generated.json","127.0.0.1","0","1","","1.14.2-x-1.0.3"],q))+" >"+q(w+"/generator-private.log")+" 2>&1")==0);
let generated=json(fs.readfile(w+"/generated.json")); let nodes=filter(generated.outbounds,o=>o.type=="vless");
assert(length(nodes)==length(filter(raw.outbounds,o=>o.type=="vless")),"generated node count changed"); let info="";
for(let i=0;i<length(nodes);i++) {
 let o=nodes[i];
 o.tag="probe-"+(i+1); delete o.remark; delete o.share_link; delete o.__forkop_description;
 info+=(i+1)+"\t"+(o.transport?.type||"tcp")+"\t"+(o.tls?.reality?.enabled?"reality":"tls")+"\n";
}
fs.writefile(w+"/nodes.tsv",info);
fs.writefile(w+"/probe.json",sprintf("%J",{log:{level:"warn"},dns:{servers:filter(base.dns.servers,s=>s.type!="fakeip"),final:"dns-server",strategy:base.dns.strategy},inbounds:[{type:"mixed",listen:"127.0.0.1",listen_port:46534}],outbounds:nodes,route:{final:nodes[0].tag,default_domain_resolver:"dns-server",auto_detect_interface:base.route.auto_detect_interface,default_mark:base.route.default_mark},experimental:{clash_api:{external_controller:"127.0.0.1:19090"}}}));
print("generated nodes=",length(nodes),"\n");
