#!/usr/bin/env python3
"""Export an actually prepared v0.39.2 host. No .env, credentials or DB volumes.
Run after install/doctor succeeds; images.tar and Ollama models are exportable.
"""
import argparse, json, os, subprocess, sys, tarfile, tempfile, shutil
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]/'installer'))
from auto_install import VERSION, read_env, compose_args, sha, atomic_write
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',default='/opt/trest-systems');p.add_argument('--output',required=True)
a=p.parse_args();root=Path(a.root).resolve();out=Path(a.output).resolve()
if out.exists() and any(out.iterdir()):raise SystemExit('ERROR: output must be a new/empty directory')
out.mkdir(parents=True,exist_ok=True);os.chmod(out,0o700)
env=read_env(root/'.env');clean=os.environ.copy()
for k in env:clean.pop(k,None)
c=compose_args(root,env)
def run(args):
 result=subprocess.run(args,env=clean,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,check=True)
 return result.stdout
conf=json.loads(run(c+['config','--format','json']))
images=sorted({s['image'] for s in conf['services'].values()})
metadata=json.loads(run(['docker','image','inspect',*images]))
if any(x.get('Architecture')!='amd64' for x in metadata):raise SystemExit('This release exporter currently supports linux/amd64 only')
subprocess.run(['docker','save','-o',str(out/'images.tar'),*images],env=clean,check=True)
container=run(c+['ps','-q','ollama']).strip()
if not container:raise SystemExit('Ollama container must be running')
with tempfile.TemporaryDirectory(prefix='trest-model-export-') as tmp:
 subprocess.run(['docker','cp',container+':/root/.ollama/models',tmp],env=clean,check=True)
 models=Path(tmp)/'models'
 if not models.is_dir():raise SystemExit('No model files exported')
 with tarfile.open(out/'ollama-models.tar','w') as t:
  t.add(models,arcname='models',recursive=True)
files={name:sha(out/name) for name in ['images.tar','ollama-models.tar']}
bundle={'version':VERSION,'architecture':'amd64','files':files,'images':[{'tags':x.get('RepoTags',[]),'id':x['Id'],'digests':x.get('RepoDigests',[])} for x in metadata], 'contains_secrets':False, 'contains_database':False}
atomic_write(out/'bundle.json',json.dumps(bundle,indent=2)+'\n')
atomic_write(out/'SHA256SUMS',''.join(d+'  '+n+'\n' for n,d in files.items()))
print('Offline bundle written:',out)
print('Combine this directory with the matching full project ZIP; keep models under release/offline/.')
