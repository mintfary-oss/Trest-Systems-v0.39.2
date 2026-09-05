#!/usr/bin/env python3
import sys, os, subprocess
from pathlib import Path
from auto_install import read_env, compose_args
root=Path(sys.argv[1]);action=sys.argv[2];env=read_env(root/".env")
args={"status":["ps","-a"],"logs":["logs","--no-color","--tail","100"],"start":["up","-d","--no-build","--pull","never"],"stop":["stop"],"restart":["restart"]}[action]
e=os.environ.copy()
for k in env:e.pop(k,None)
e.pop("COMPOSE_FILE",None)
sys.exit(subprocess.call(compose_args(root,env)+args+sys.argv[3:],env=e))
