#!/usr/bin/env python3
"""Package the complete source + prebuilt release. Never implies runtime readiness.
Does not fetch Docker images/model weights or include credentials/runtime data.
"""
import argparse,datetime,hashlib,json,os,shutil,stat,subprocess,tempfile,zipfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
VERSION='0.39.2'
REQUIRED=['trest-api','trest-worker','trest','trestctl','trest-install','trest-installer','generate-dxf']
def digest(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
 return h.hexdigest()
def write(p,obj):
 p.parent.mkdir(parents=True,exist_ok=True)
 p.write_text(json.dumps(obj,ensure_ascii=False,indent=2)+'\n')
def include(p):
 rel=p.relative_to(ROOT)
 if any(x in {'.git','node_modules','__pycache__','.pytest_cache','.mypy_cache','.next','dist','runtime','installation-reports'} for x in rel.parts):return False
 if p.suffix in {'.pyc','.pyo','.tmp','.ttf','.otf','.woff','.woff2'}:return False
 if p.name in {'.env','env.private','admin.txt','postgres.dump'}:return False
 if p.name.endswith(('.zip','.zip.sha256','.dump')):return False
 if rel.parts[:2]==('verification','dry-run'):return False
 if rel.parts[:2]==('release','offline'):return False
 return p.is_file()
def prepare_manifest():
 payload=[]
 paths=[ROOT/'release/bin/linux/amd64'/n for n in REQUIRED]+[ROOT/'release/bin/windows/amd64/trest-install.exe']
 for p in paths:
  if not p.is_file():raise RuntimeError('Required binary missing: '+str(p))
  if not (p.read_bytes()[:4]==b'\x7fELF' or p.read_bytes()[:2]==b'MZ'):raise RuntimeError('Not an executable: '+str(p))
  payload.append({'path':str(p.relative_to(ROOT)),'size':p.stat().st_size,'sha256':digest(p)})
 (ROOT/'release/bin/SHA256SUMS').write_text(''.join(b['sha256']+'  '+b['path']+'\n' for b in payload))
 try:goversion=subprocess.check_output(['go','version'],text=True).strip()
 except (OSError,subprocess.CalledProcessError):goversion='compiler not available to packager; see binary go build metadata'
 info={'release':'0.39.2','classification':'release candidate','generated_at_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'build_host_go':goversion,'target':'linux/amd64; Windows/amd64 bootstrap launcher only','strategy':'prebuilt Go; Docker images/model weights NOT included','binaries':payload,'database_driver':'inherited pgx-compatible API using system libpq; NOT full upstream pgx','required_runtime_packages':['glibc >= 2.34','libpq5 and dependencies','ca-certificates','tzdata'],'docker_runtime_verified':False,'windows_runtime_verified':False}
 write(ROOT/'release/bin/BUILD_INFO.json',info)
 names={}
 for p in sorted(ROOT.rglob('*')):
  if not include(p):continue
  rel=p.relative_to(ROOT)
  if p.is_symlink():raise RuntimeError('Source symlink refused: '+str(rel))
  if rel.parts[0]=='verification' or rel.parts[:2]==('deployments','generated') or str(rel)=='release/package-manifest.json':continue
  names[str(rel)]=digest(p)
 write(ROOT/'release/package-manifest.json',{'version':VERSION,'architectures':['amd64'],'files':names,'scope':'package file integrity; not external signature, not runtime acceptance','excluded_mutable':['verification/','deployments/generated/','.env','release/offline/']})
 return names
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--output',default=str(ROOT/'dist'));p.add_argument('--manifest-only',action='store_true');a=p.parse_args()
 names=prepare_manifest()
 if a.manifest_only:print('Manifest files:',len(names));return
 dest=Path(a.output).resolve();dest.mkdir(parents=True,exist_ok=True)
 archive=dest/f'Trest-Systems-v{VERSION}-FULL-FIXED.zip'
 exclude_dest = dest.is_relative_to(ROOT)
 entries=[p for p in sorted(ROOT.rglob('*')) if include(p) and p!=archive and not (exclude_dest and p.resolve().is_relative_to(dest))]
 # An output equal to source root would make include/exclusion ambiguous.
 if dest==ROOT:raise SystemExit('Choose an output directory different from source root')
 with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED,compresslevel=6) as z:
  for p in entries:
   rel=p.relative_to(ROOT)
   if p.is_symlink():raise RuntimeError('Symlink refused: '+str(rel))
   info=zipfile.ZipInfo.from_file(p,str(Path('Trest-Systems-v'+VERSION)/rel))
   mode=stat.S_IMODE(p.stat().st_mode)
   if p.suffix=='.sh' or rel.parts[:4]==('release','bin','linux','amd64'):
    mode |= 0o111
   info.create_system=3
   info.external_attr=(stat.S_IFREG|mode)<<16
   z.writestr(info,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=6)
 with zipfile.ZipFile(archive) as z:
  bad=z.testzip()
  if bad:raise RuntimeError('ZIP CRC error: '+bad)
  if len(z.namelist())!=len(set(z.namelist())):raise RuntimeError('Duplicate ZIP paths')
 (dest/(archive.name+'.sha256')).write_text(digest(archive)+'  '+archive.name+'\n')
 print(archive);print('Files:',len(entries));print('Bytes:',archive.stat().st_size);print('SHA256:',digest(archive));print('Docker images/models not included; runtime verification not inferred.')
if __name__=='__main__':main()
