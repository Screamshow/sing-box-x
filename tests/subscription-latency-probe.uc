// Run only with Forkop stopped and the isolated candidate/API started.
let fs=require("fs"); let w=ARGV[0]; let results=[];
for(let round in [1,2,3]) {
 let lines=split(trim(fs.readfile(w+"/nodes.tsv")),"\n");
 for(let line in lines) {
  let fields=split(line,"\t"),n=fields[0],path=w+"/delay-"+round+"-"+n+".json";
  let rc=system("curl -fsS --max-time 10 -G http://127.0.0.1:19090/proxies/probe-"+n+"/delay --data-urlencode 'url=https://www.gstatic.com/generate_204' --data-urlencode 'timeout=8000' -o "+path+" 2>/dev/null");
  let v=json(fs.readfile(path)||"{}"); let delay=v?.delay; let ok=rc==0&&type(delay)=="int"&&delay>=0;
  push(results,{round,id:+n,transport:fields[1],security:fields[2],ok,delay:ok?delay:null});
 }
 let current=filter(results,r=>r.round==round); print("round=",round," passed=",length(filter(current,r=>r.ok))," failed=",length(filter(current,r=>!r.ok)),"\n");
 fs.writefile(w+"/anonymous-results.json",sprintf("%J",results));
}
