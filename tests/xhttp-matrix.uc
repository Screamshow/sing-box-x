// Private TLS/REALITY keys are generated on the VM, never fixtures.
let fs = require("fs");
let w = ARGV[0]; let security = ARGV[1]; let mode = ARGV[2]; let download = ARGV[3] == "1";
let pem = fs.readfile(w + "/tls.pem");
function block(label) {
    let a = index(pem,"-----BEGIN "+label+"-----"), b = index(pem,"-----END "+label+"-----");
    return substr(pem,a,b+length("-----END "+label+"-----")-a);
}
function write(name,obj) { fs.writefile(w+"/"+name,sprintf("%J",obj)); }
let keys = fs.readfile(w+"/reality.keys");
function key(labels) {
    for (let line in split(keys,"\n")) for (let label in labels)
        if (index(line,label) == 0) return trim(substr(line,length(label)));
    die("test key missing\n");
}
let cert = block("CERTIFICATE"), privateTLS = block("PRIVATE KEY");
write("target.json",{log:{level:"error"},inbounds:[{type:"http",listen:"127.0.0.1",listen_port:18443,
    tls:{enabled:true,certificate:[cert],key:[privateTLS]}}],outbounds:[{type:"direct"}]});
let uuid = "00000000-0000-4000-8000-000000000001";
let stream = {network:"xhttp",security:security=="reality"?"reality":security=="plain"?"none":"tls",
    xhttpSettings:{path:"/probe",mode:"auto"}};
if (mode=="tcp") { stream.network="tcp"; delete stream.xhttpSettings; }
if (security=="reality") stream.realitySettings={target:"127.0.0.1:18443",serverNames:["example.test"],
    privateKey:key(["PrivateKey:","Private key:"]),shortIds:["0123456789abcdef"]};
else if (security!="plain") stream.tlsSettings={alpn:security=="h1"?["http/1.1"]:["h2"],
    certificates:[{certificate:split(cert,"\n"),key:split(privateTLS,"\n")}]};
let inbounds=[{listen:"127.0.0.1",port:19443,protocol:"vless",settings:{clients:[{id:uuid}],decryption:"none"},streamSettings:stream}];
write("server.json",{log:{loglevel:"error"},inbounds,outbounds:[{protocol:"freedom",settings:{finalRules:[{action:"allow",ip:["127.0.0.1"],port:18090}]}}]});
let tls = {enabled:true,server_name:"example.test",insecure:true,alpn:security=="h1"?["http/1.1"]:["h2"]};
if(security=="reality") {
    tls.insecure=false;
    tls.utls={enabled:true,fingerprint:"firefox"};
    tls.reality={enabled:true,public_key:key(["Password (PublicKey):","Public key:"]),short_id:"0123456789abcdef"};
}
let transport={type:"xhttp",path:"/probe",mode};
if(download) transport.download={server:"127.0.0.1",server_port:19443,path:"/probe",tls};
let outbound={type:"vless",tag:"proxy",server:"127.0.0.1",server_port:19443,uuid,transport};
if(security!="plain") outbound.tls=tls;
if(mode=="tcp") delete outbound.transport;
write("client.json",{log:{level:"error"},inbounds:[{type:"mixed",listen:"127.0.0.1",listen_port:19444}],outbounds:[outbound],route:{final:"proxy"}});
