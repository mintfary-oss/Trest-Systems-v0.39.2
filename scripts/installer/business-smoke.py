#!/usr/bin/env python3
"""Opt-in business smoke on a DISPOSABLE installation; never downs a stack.
Creates test users/projects. Not comprehensive E2E or production acceptance.
"""
import argparse,json,secrets,sys,urllib.request,urllib.error
from pathlib import Path
from auto_install import read_env
p=argparse.ArgumentParser(description=__doc__);p.add_argument('root');p.add_argument('--confirm-disposable',action='store_true');a=p.parse_args()
if not a.confirm_disposable:raise SystemExit('Refusing test data writes. Use a disposable install and --confirm-disposable.')
e=read_env(Path(a.root)/'.env');base=e['PUBLIC_URL'];email='smoke-'+secrets.token_hex(10)+'@test.invalid';password=secrets.token_urlsafe(24)
def req(method,path,status,body=None,token=None):
 headers={'Content-Type':'application/json'}
 if token:headers['Authorization']='Bearer '+token
 q=urllib.request.Request(base+path,data=json.dumps(body).encode() if body is not None else None,headers=headers,method=method)
 try:
  with urllib.request.urlopen(q,timeout=30) as r:code=r.status;data=r.read()
 except urllib.error.HTTPError as r:code=r.code;data=r.read()
 if code!=status:raise RuntimeError(f'{method} {path}: expected {status}, got {code}')
 print('PASS',method,path,code)
 return json.loads(data) if data else None
req('GET','/ready',200)
req('GET','/api/v1/projects',401)
req('POST','/api/v1/auth/register',201,{'email':email,'name':'Smoke Test','password':password})
token=req('POST','/api/v1/auth/login',200,{'email':email,'password':password})['access_token']
req('GET','/api/v1/me',200,token=token)
req('POST','/api/v1/projects',201,{'name':'Disposable smoke project','location':{},'parameters':{}},token)
req('GET','/api/v1/projects',200,token=token)
req('GET','/marketplace-api/health',200)
req('GET','/marketplace-api/api/products',200)
print('Selected business smoke: PASS (not full product E2E). Test records remain in the disposable database.')
