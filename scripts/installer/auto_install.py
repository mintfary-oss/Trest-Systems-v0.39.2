#!/usr/bin/env python3
"""Trest 0.39.2. Standard-library-only, non-destructive Linux installer.

One migration runner, preserved .env, explicit ownership, no production down -v.
Runtime acceptance is SMOKE + optional isolated DB restore, not full product E2E.
"""
from __future__ import annotations
import argparse
import datetime as dt
import fcntl
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import platform
import re
import secrets
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request

VERSION = "0.39.2"
SOURCE = Path(__file__).resolve().parents[2]
SECRET_KEYS = {"POSTGRES_PASSWORD", "REDIS_PASSWORD", "MINIO_ROOT_PASSWORD", "JWT_SECRET", "SECRET_KEY", "ADMIN_PASSWORD", "WEBUI_SECRET_KEY", "WEBUI_ADMIN_PASSWORD", "DATABASE_URL", "MARKETPLACE_DATABASE_URL", "REDIS_URL", "MINIO_SECRET_KEY"}
REQUIRED_SERVICES = ("postgres", "redis", "minio", "api", "worker", "marketplace-api", "web", "nginx", "edge", "ollama", "super-sistema-webui")


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat()


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for b in iter(lambda: f.read(1024 * 1024), b""):
            h.update(b)
    return h.hexdigest()


def read_env(path: Path) -> dict[str, str]:
    """Parse single-line Docker dotenv assignments without executing shell code.

    Existing unknown lines/comments are preserved by update_env. Unsupported
    multiline/escaped constructs fail explicitly instead of silently rewriting.
    """
    result = {}
    if not path.exists():
        return result
    for number, raw in enumerate(path.read_text(encoding="utf-8-sig").splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].lstrip()
        m = re.match(r"^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$", line)
        if not m:
            raise ValueError(f"Unsupported .env syntax at line {number}; file unchanged")
        key, val = m.groups()
        if val.startswith("'"):
            match = re.fullmatch(r"'((?:\\.|[^'\\])*)'\s*(?:#.*)?", val)
            if not match:
                raise ValueError(f"Unsupported quoted .env value at line {number}")
            val = match[1].replace("\\'", "'").replace("\\\\", "\\")
        elif val.startswith('"'):
            match = re.fullmatch(r'"((?:\\.|[^"\\])*)"\s*(?:#.*)?', val)
            if not match:
                raise ValueError(f"Unsupported quoted .env value at line {number}")
            val = json.loads('"' + match[1] + '"')
        else:
            val = re.split(r"\s+#", val, 1)[0].strip()
        if key in result:
            raise ValueError(f"Duplicate .env key {key}; review rather than guessing")
        result[key] = val
    return result


def env_value(value: str) -> str:
    if "\n" in value or "\r" in value or "\x00" in value:
        raise ValueError("Newline/NUL in environment value is not supported")
    if re.fullmatch(r"[A-Za-z0-9_:/.,@%?&=+; -]*", value) and "$" not in value and "#" not in value:
        return "'" + value + "'"
    return "'" + value.replace("\\", "\\\\").replace("'", "\\'") + "'"


def atomic_write(path: Path, content: str, mode: int = 0o600):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp = tempfile.mkstemp(prefix="." + path.name, dir=path.parent)
    try:
        os.fchmod(fd, mode)
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(content)
            f.flush()
            os.fsync(f.fileno())
        os.replace(temp, path)
    finally:
        if os.path.exists(temp):
            os.unlink(temp)


def update_env(path: Path, updates: dict[str, str]):
    # Validate before changing any bytes.
    read_env(path)
    old = path.read_text(encoding="utf-8-sig").splitlines() if path.exists() else ["# Trest Systems configuration. PRIVATE: do not commit or share."]
    seen = set()
    lines = []
    for line in old:
        m = re.match(r"^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=", line)
        if m and m[1] in updates:
            seen.add(m[1]); lines.append(m[1] + "=" + env_value(str(updates[m[1]])))
        else:
            lines.append(line)
    for key, val in updates.items():
        if key not in seen:
            lines.append(key + "=" + env_value(str(val)))
    atomic_write(path, "\n".join(lines) + "\n")


def compose_args(root: Path, env: dict[str, str]) -> list[str]:
    args = ["docker", "compose", "--project-name", env.get("COMPOSE_PROJECT_NAME", "trest"), "--env-file", str(root / ".env"), "-f", str(root / "deployments/docker-compose.yml")]
    if env.get("TLS_MODE", "off") != "off":
        args += ["-f", str(root / "deployments/tls-ports.yml")]
    return args


def owned_port(containers: list[dict], project: str, service: str, port: int) -> bool:
    for c in containers:
        labels = c.get("Config", {}).get("Labels", {}) or {}
        if labels.get("com.docker.compose.project") != project or labels.get("com.docker.compose.service") != service:
            continue
        ports = c.get("NetworkSettings", {}).get("Ports", {}) or {}
        if any(str(port) == binding.get("HostPort") for bindings in ports.values() for binding in (bindings or [])):
            return True
    return False


class Installer:
    def __init__(self, args):
        self.a = args
        self.root = Path(args.install_dir).expanduser().resolve()
        if self.root == Path("/") or self.root in {Path("/opt"), Path("/root"), Path("/home"), Path("/usr")}:
            raise ValueError("Unsafe installation directory")
        self.env = read_env(self.root / ".env")
        self.private = []
        self.refresh_secrets()
        self.started = now()
        self.tag = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ") + "-" + secrets.token_hex(3)
        self.steps, self.warnings = [], []
        self.errors = []
        self.logs = []
        self.arch = {"x86_64": "amd64", "aarch64": "arm64"}.get(platform.machine(), platform.machine())
        self.public = ""
        self.state = Path(args.state_dir)
        self.backup_path = None
        self.backup_index = 0
        self.new_install = not (self.root / ".env").exists()

    def refresh_secrets(self):
        self.private = sorted({v for k, v in self.env.items() if k in SECRET_KEYS and len(v) >= 4}, key=len, reverse=True)

    def redact(self, text: str) -> str:
        for secret in self.private:
            text = text.replace(secret, "[REDACTED]")
        text = re.sub(r"(postgres(?:ql)?|redis)://[^\s@]+@", r"\1://[REDACTED]@", text)
        return text

    def info(self, text):
        line = self.redact(str(text))
        print(line, flush=True); self.logs.append(line)

    def warn(self, text):
        self.warnings.append(self.redact(text)); self.info("[WARN] " + text)

    def stage(self, title, fn):
        record = {"name": title, "started_at": now(), "status": "RUNNING"}
        self.steps.append(record); self.info("=== " + title + " ===")
        try:
            result = fn()
            record["status"] = "PASS"; self.info("[PASS] " + title)
            return result
        except Exception as e:
            record["status"] = "FAIL"; record["error"] = self.redact(str(e)); raise
        finally:
            record["finished_at"] = now()

    def run(self, args, *, data=None, quiet=False, check=True, timeout=900, cwd=None):
        # Inherited shell variables must not override the installation's .env.
        env = os.environ.copy()
        for key in self.env:
            env.pop(key, None)
        env.pop("COMPOSE_FILE", None); env.pop("COMPOSE_PROFILES", None)
        env["COMPOSE_ANSI"] = "never"
        if not quiet:
            self.info("$ " + " ".join(str(x) for x in args))
        try:
            p = subprocess.run([str(x) for x in args], input=data, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env, cwd=cwd or (self.root if self.root.exists() else SOURCE), timeout=timeout)
        except subprocess.TimeoutExpired:
            raise RuntimeError("Command timeout: " + str(args[0])) from None
        output = self.redact(p.stdout + p.stderr)
        if not quiet and output:
            self.info(output[-20000:])
        if p.returncode and check:
            raise RuntimeError(f"Command exited {p.returncode}: " + output[-3000:])
        return p

    def dc(self, *args, **kwargs):
        return self.run(compose_args(self.root, self.env) + list(args), **kwargs)

    def config(self):
        return json.loads(self.dc("config", "--format", "json", quiet=True).stdout)

    def package_check(self):
        manifest = SOURCE / "release/package-manifest.json"
        if not manifest.is_file():
            raise RuntimeError("release/package-manifest.json missing")
        meta = json.loads(manifest.read_text())
        if self.arch not in meta["architectures"]:
            raise RuntimeError(f"No verified binaries for {self.arch}; supported: {meta['architectures']}")
        for name, digest in meta["files"].items():
            p = SOURCE / name
            if not p.resolve().is_relative_to(SOURCE) or p.is_symlink() or not p.is_file():
                raise RuntimeError("Missing/unsafe package file: " + name)
            if sha(p) != digest:
                raise RuntimeError("Package checksum mismatch: " + name)
        self.info(f"Package SHA-256 checks: {len(meta['files'])} files")

    def system_check(self):
        if platform.system() != "Linux":
            raise RuntimeError("Run install.sh on Linux; Windows uses install.ps1 + WSL2")
        if not self.a.dry_run and os.geteuid() != 0:
            raise RuntimeError("Run sudo ./install.sh")
        mem = int(re.search(r"MemTotal:\s+(\d+)", Path('/proc/meminfo').read_text())[1]) * 1024
        probe = self.root
        while not probe.exists():
            probe = probe.parent
        free = shutil.disk_usage(probe).free
        self.info(f"CPU={os.cpu_count()} RAM={mem//2**20} MiB disk_free={free//2**30} GiB arch={self.arch}")
        if mem < 4 * 2**30 or free < 8 * 2**30:
            raise RuntimeError("Minimum: 4 GiB RAM and 8 GiB free disk; AI images/models may require substantially more")
        if free < 25 * 2**30:
            self.warn("Below recommended 25 GiB free disk for images, models and backups")

    def prerequisites(self):
        needed = [x for x in ("python3", "curl", "openssl", "tar", "ss") if not shutil.which(x)]
        if self.a.offline:
            if needed or not shutil.which("docker"):
                raise RuntimeError("Offline install requires Python3/Docker/Compose and base system tools already installed")
            return
        if needed:
            if not shutil.which("apt-get"):
                raise RuntimeError("Automatic prerequisites supported on Debian/Ubuntu; install Docker/Python3/curl/openssl/tar/iproute2 first")
            self.run(["apt-get", "-o", "Acquire::Retries=3", "-o", "Acquire::http::Timeout=30", "update"])
            self.run(["apt-get", "install", "-y", "--no-install-recommends", "python3", "ca-certificates", "curl", "openssl", "tar", "iproute2"])
        if not shutil.which("docker"):
            osinfo = dict(line.split("=", 1) for line in Path("/etc/os-release").read_text().splitlines() if "=" in line)
            distro = osinfo.get("ID", "").strip('"'); code = osinfo.get("VERSION_CODENAME", "").strip('"')
            if distro not in {"debian", "ubuntu"} or not re.fullmatch(r"[a-z]+", code):
                raise RuntimeError("Install Docker Engine and Compose v2 for this distribution first")
            self.run(["apt-get", "update"])
            self.run(["apt-get", "install", "-y", "ca-certificates", "curl"])
            Path("/etc/apt/keyrings").mkdir(exist_ok=True)
            self.run(["curl", "-fSL", "--connect-timeout", "10", "--max-time", "90", f"https://download.docker.com/linux/{distro}/gpg", "-o", "/etc/apt/keyrings/trest-docker.asc"])
            os.chmod('/etc/apt/keyrings/trest-docker.asc', 0o644)
            aptarch = self.run(["dpkg", "--print-architecture"], quiet=True).stdout.strip()
            atomic_write(Path('/etc/apt/sources.list.d/trest-docker.list'), f"deb [arch={aptarch} signed-by=/etc/apt/keyrings/trest-docker.asc] https://download.docker.com/linux/{distro} {code} stable\n", 0o644)
            self.run(["apt-get", "update"])
            self.run(["apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"])
        if shutil.which("systemctl"):
            self.run(["systemctl", "enable", "--now", "docker"])
        self.run(["docker", "info"], quiet=True)
        v = self.run(["docker", "compose", "version", "--short"], quiet=True).stdout.strip()
        match = re.search(r"(\d+)\.(\d+)\.(\d+)", v)
        if not match or tuple(map(int, match.groups())) < (2, 24, 0):
            raise RuntimeError("Docker Compose >=2.24 required; existing Docker was not removed or replaced")
        self.info("Docker Compose " + v)

    def existing_containers(self):
        ids = self.run(["docker", "ps", "-aq"], quiet=True).stdout.split()
        return json.loads(self.run(["docker", "inspect", *ids], quiet=True).stdout) if ids else []

    def configure(self):
        old = dict(self.env); v = dict(old)
        v.setdefault("COMPOSE_PROJECT_NAME", "trest")
        v.setdefault("POSTGRES_USER", "trest"); v.setdefault("POSTGRES_DB", "trest")
        if not re.fullmatch(r"[a-z0-9][a-z0-9_-]*", v["COMPOSE_PROJECT_NAME"]):
            raise RuntimeError("Invalid COMPOSE_PROJECT_NAME")
        for k in ("POSTGRES_USER", "POSTGRES_DB"):
            if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]{0,62}", v[k]):
                raise RuntimeError("Unsupported PostgreSQL identifier: " + k)
        existing = self.existing_containers()
        own_db = [c for c in existing if (c.get('Config',{}).get('Labels') or {}).get('com.docker.compose.project') == v['COMPOSE_PROJECT_NAME'] and (c.get('Config',{}).get('Labels') or {}).get('com.docker.compose.service') == 'postgres']
        if own_db and not old.get("POSTGRES_PASSWORD"):
            raise RuntimeError("Existing PostgreSQL found but .env password missing: recover config backup; password NOT regenerated")
        for k, size in {"POSTGRES_PASSWORD":32,"REDIS_PASSWORD":32,"MINIO_ROOT_PASSWORD":32,"JWT_SECRET":48,"SECRET_KEY":48,"ADMIN_PASSWORD":24,"WEBUI_SECRET_KEY":48,"WEBUI_ADMIN_PASSWORD":24}.items():
            if not v.get(k):
                v[k] = secrets.token_hex(size)
        v.setdefault("ADMIN_EMAIL", "admin@trest.local")
        v.setdefault("MINIO_ROOT_USER", "trest")
        v.setdefault("CHAT_MODEL", "llama3.2:3b")
        v.setdefault("EMBEDDING_MODEL", "nomic-embed-text")
        for model in (v['CHAT_MODEL'], v['EMBEDDING_MODEL']):
            if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_./:-]{0,150}", model):
                raise RuntimeError("Invalid model name")
        domain = self.a.domain if self.a.domain is not None else v.get("DOMAIN", "")
        tls = self.a.tls if self.a.tls is not None else v.get("TLS_MODE", "auto")
        if tls == "auto" and not domain:
            tls = "off"
        if domain and not re.fullmatch(r"(?:[A-Za-z0-9][A-Za-z0-9-]*\.)+[A-Za-z]{2,63}", domain):
            raise RuntimeError("Invalid domain name")
        email = self.a.email if self.a.email is not None else v.get("ACME_EMAIL", "")
        if tls == "auto" and not re.fullmatch(r"[^\s@]+@[^\s@]+\.[^\s@]+", email):
            raise RuntimeError("Trusted TLS requires --domain and --email")
        v.update(DOMAIN=domain,TLS_MODE=tls,ACME_EMAIL=email,TARGETARCH=self.arch,TREST_VERSION=VERSION)
        ports = [("HTTP_PORT", self.a.http_port or v.get('HTTP_PORT','80'), "edge", "0.0.0.0"), ("OLLAMA_PORT", v.get("OLLAMA_PORT","11435"), "ollama", "127.0.0.1"), ("WEBUI_PORT", v.get("WEBUI_PORT","3001"), "super-sistema-webui", "127.0.0.1"), ("MINIO_PORT", v.get("MINIO_PORT","9000"), "minio", "127.0.0.1"), ("MINIO_CONSOLE_PORT", v.get("MINIO_CONSOLE_PORT","9001"), "minio", "127.0.0.1")]
        if tls != "off":
            ports.append(("HTTPS_PORT", self.a.https_port or v.get('HTTPS_PORT','443'), "edge", "0.0.0.0"))
        used = set()
        for key, value, service, bind in ports:
            port = int(value)
            if not 1 <= port <= 65535:
                raise RuntimeError("Invalid port: " + key)
            def free(p):
                if p in used:
                    return False
                if owned_port(existing,v["COMPOSE_PROJECT_NAME"],service,p):
                    return True
                with socket.socket(socket.AF_INET,socket.SOCK_STREAM) as s:
                    try: s.bind((bind,p));return True
                    except OSError:return False
            if not free(port):
                if self.a.port_policy == "fail" or (tls != "off" and service == "edge"):
                    raise RuntimeError(f"Port {port} is owned by another service; no container was stopped. Choose --http-port or --port-policy auto")
                alternative = next((p for p in range(max(8088,port+1), min(65536,max(8088,port+1)+500)) if free(p)), None)
                if alternative is None:
                    raise RuntimeError("No available port for " + key)
                self.warn(f"Port {port} occupied; {key}={alternative}. Existing service left unchanged")
                port = alternative
            v[key] = str(port);used.add(port)
        v.setdefault("HTTPS_PORT","443")
        if tls == "auto" and (v["HTTP_PORT"]!="80" or v["HTTPS_PORT"]!="443"):
            raise RuntimeError("Automatic public TLS requires external 80 and 443")
        host = domain or self.a.public_host or v.get("PUBLIC_HOST", "")
        if not host:
            try:
                with socket.socket(socket.AF_INET,socket.SOCK_DGRAM) as s:
                    s.connect(("1.1.1.1",80));host = s.getsockname()[0]
            except OSError:host = "127.0.0.1"
        if not domain:
            ipaddress.ip_address(host)
        host_url = '['+host+']' if ':' in host else host
        protocol, port = ("http",v['HTTP_PORT']) if tls=='off' else ("https",v['HTTPS_PORT'])
        self.public = f"{protocol}://{host_url}" + ("" if port==('80' if tls=='off' else '443') else ':'+port)
        quote = lambda x: urllib.parse.quote(x,safe="")
        uri = f"postgresql://{quote(v['POSTGRES_USER'])}:{quote(v['POSTGRES_PASSWORD'])}@postgres:5432/{quote(v['POSTGRES_DB'])}"
        v.update(DATABASE_URL=uri+"?sslmode=disable",MARKETPLACE_DATABASE_URL=uri,REDIS_URL=f"redis://:{quote(v['REDIS_PASSWORD'])}@redis:6379/0",PUBLIC_URL=self.public,PUBLIC_HOST=host,WEBUI_AUTH="true",DEFAULT_MODELS=v['CHAT_MODEL'])
        self.root.mkdir(parents=True,exist_ok=True)
        self.state.mkdir(parents=True,exist_ok=True)
        if old:
            backup=self.state/'backups'/(self.tag+'-config');backup.mkdir(parents=True,exist_ok=True)
            shutil.copy2(self.root/'.env',backup/'env.private');os.chmod(backup/'env.private',0o600)
        update_env(self.root/'.env',v)
        self.env=read_env(self.root/'.env');self.refresh_secrets()
        creds=f"Trest {VERSION}\nURL: {self.public}\nAdmin email: {v['ADMIN_EMAIL']}\nInitial marketplace/core password: {v['ADMIN_PASSWORD']}\nInitial WebUI password: {v['WEBUI_ADMIN_PASSWORD']}\nExisting users are never reset. Change initial passwords after first login.\nWebUI is loopback-only; use an SSH tunnel.\n"
        atomic_write(self.state/'credentials/admin.txt',creds)
        self.info("Credentials: " + str(self.state/'credentials/admin.txt') + " (0600); values not printed")

    def copy_project(self):
        if self.root == SOURCE:
            return
        if self.root.is_relative_to(SOURCE):
            raise RuntimeError("Installation directory must not be inside source tree")
        for p in SOURCE.iterdir():
            if p.name in {'.git','.env','.trest','runtime','installation-reports','node_modules'} or p.name.endswith('.bak'):
                continue
            dest=self.root/p.name
            if p.is_symlink():raise RuntimeError("Source symlink refused: "+p.name)
            if p.is_dir():shutil.copytree(p,dest,dirs_exist_ok=True,ignore=shutil.ignore_patterns('__pycache__','node_modules','.git','.env','*.pyc'))
            else:shutil.copy2(p,dest)
        for p in (self.root/'release/bin/linux'/self.arch).iterdir():
            p.chmod(0o755)

    def render(self):
        tls=self.env['TLS_MODE'];domain=self.env['DOMAIN']
        if tls == 'auto':
            address=domain; extra='  email '+self.env['ACME_EMAIL']+'\n'
        elif tls == 'internal':
            address=domain or 'localhost';extra=''
        else:address=':80';extra='  auto_https off\n'
        internal='  tls internal\n' if tls=='internal' else ''
        content='{\n  admin off\n  persist_config off\n'+extra+'}\n'+address+' {\n'+internal+'  encode gzip zstd\n  header X-Trest-Release "'+VERSION+'"\n  header X-Content-Type-Options nosniff\n  header Referrer-Policy strict-origin-when-cross-origin\n  reverse_proxy nginx:80\n}\nhttp://127.0.0.1:8088 {\n  respond /health "ok" 200\n}\n'
        atomic_write(self.root/'deployments/generated/caddy/Caddyfile',content,0o644)
        # This is loopback inside the container, NOT an externally published admin port.
        if tls=='off':self.warn("HTTP mode: no transport encryption. Use test data; configure trusted HTTPS before production credentials")
        if tls=='internal':self.warn("Internal CA is not trusted by remote browsers until its root is installed")

    def configure_firewall(self):
        if self.a.no_firewall:
            self.warn('Firewall rule updates explicitly skipped');return
        ports=[self.env['HTTP_PORT']]
        if self.env['TLS_MODE']!='off':ports.append(self.env['HTTPS_PORT'])
        if shutil.which('ufw') and 'Status: active' in self.run(['ufw','status'],quiet=True,check=False).stdout:
            for port in ports:self.run(['ufw','allow',port+'/tcp','comment','Trest HTTP/TLS'])
        elif shutil.which('firewall-cmd') and self.run(['firewall-cmd','--state'],quiet=True,check=False).returncode==0:
            for port in ports:
                self.run(['firewall-cmd','--add-port='+port+'/tcp'])
                self.run(['firewall-cmd','--permanent','--add-port='+port+'/tcp'])
        else:self.warn('No active supported host firewall detected; none enabled/reset. Cloud firewall and external reachability require separate verification')

    def images(self):
        bundle=Path(self.a.image_bundle).resolve() if self.a.image_bundle else self.root/'release/offline'
        meta=bundle/'bundle.json'
        if meta.exists():
            spec=json.loads(meta.read_text())
            if spec.get('version')!=VERSION or spec.get('architecture')!=self.arch:
                raise RuntimeError('Offline bundle version/architecture mismatch')
            for name,digest in spec['files'].items():
                p=bundle/name
                if not p.resolve().is_relative_to(bundle) or sha(p)!=digest:raise RuntimeError('Offline bundle checksum mismatch')
            self.run(['docker','load','-i',str(bundle/'images.tar')],timeout=1800)
        elif self.a.offline:
            raise RuntimeError('OFFLINE bundle absent. This source+binary package does not contain Docker images. Use scripts/release/export-offline.py on a prepared host')
        conf=self.config()
        for name,s in conf['services'].items():
            image=s.get('image')
            if not image:raise RuntimeError('Service missing image name: '+name)
            exists=self.run(['docker','image','inspect',image],quiet=True,check=False).returncode==0
            if self.a.offline:
                if not exists:raise RuntimeError('Offline image missing: '+image)
            elif 'build' in s and (not exists or self.a.rebuild):
                self.dc('build',name,timeout=1800)
            elif not exists:
                self.run(['docker','pull',image],timeout=1800)
        self.bundle=bundle

    def start_services(self,*services):
        self.dc('up','-d','--no-build','--pull','never','--wait','--wait-timeout',str(self.a.wait_timeout),*services,timeout=self.a.wait_timeout+60)

    def psql(self,sql,database=None):
        return self.dc('exec','-T','postgres','psql','-X','-qAt','-v','ON_ERROR_STOP=1','-U',self.env['POSTGRES_USER'],'-d',database or self.env['POSTGRES_DB'],'-f','-',data=sql,quiet=True)

    def password_check(self):
        # TCP password check, not pg_isready or local trust authentication.
        p=self.dc('exec','-T','postgres','sh','-ec','PGPASSWORD="$POSTGRES_PASSWORD" psql -X -w -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atqc "SELECT 1"',quiet=True,check=False)
        if p.returncode:
            if not self.a.sync_postgres_password:
                raise RuntimeError('PostgreSQL password differs from saved .env. Recover the previous config or explicitly --sync-postgres-password. Volume was not deleted')
            passwd=self.env['POSTGRES_PASSWORD'].replace("'","''")
            self.psql('SET standard_conforming_strings=on; ALTER ROLE "'+self.env['POSTGRES_USER']+'" PASSWORD \''+passwd+"';\n")
            self.dc('exec','-T','postgres','sh','-ec','PGPASSWORD="$POSTGRES_PASSWORD" psql -X -w -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atqc "SELECT 1"',quiet=True)
        self.info('PostgreSQL TCP authentication: PASS')

    def check_legacy_shop(self):
        # Do not quietly orphan the historical shared-schema shop when upgrading.
        legacy=self.psql("SELECT to_regclass('public.products') IS NOT NULL AND to_regclass('marketplace.products') IS NULL;\n").stdout.strip()
        if legacy=='t':
            raise RuntimeError('Legacy marketplace in public schema detected. Database backup created. Automatic shared-table conversion is deliberately blocked; see docs/install/UPGRADE_RU.md. Do not delete the database')

    def migrate(self):
        self.dc('run','--rm','--no-deps','-T','api','--migrate-only',timeout=900)
        self.dc('run','--rm','--no-deps','-T','api','--verify-migrations',timeout=120)

    def models(self):
        if self.a.skip_models:
            self.warn('Model download/inference checks SKIPPED by --skip-models. AI readiness not established')
            return
        archive=self.bundle/'ollama-models.tar'
        if archive.exists():
            # Archive content paths are checked before piping to tar inside container.
            import tarfile
            with tarfile.open(archive) as t:
                for m in t.getmembers():
                    p=Path(m.name)
                    if p.is_absolute() or '..' in p.parts or m.issym() or m.islnk() or m.isdev() or not (m.name=='models' or m.name.startswith('models/')):
                        raise RuntimeError('Unsafe path/type in models archive')
            ids=self.dc('ps','-q','ollama',quiet=True).stdout.strip()
            # docker cp archives a regular file, followed by in-container extraction.
            self.run(['docker','cp',str(archive),ids+':/tmp/trest-models.tar'])
            self.dc('exec','-T','ollama','sh','-ec','tar -xf /tmp/trest-models.tar -C /root/.ollama && rm /tmp/trest-models.tar')
        listed=self.dc('exec','-T','ollama','ollama','list',quiet=True).stdout
        for model in (self.env['EMBEDDING_MODEL'],self.env['CHAT_MODEL']):
            canonical=model if ':' in model else model+':latest'
            if canonical not in [line.split()[0] for line in listed.splitlines()[1:] if line.strip()]:
                if self.a.offline:raise RuntimeError('Offline model absent: '+model)
                self.dc('exec','-T','ollama','ollama','pull',model,timeout=1800)
        self.dc('exec','-T','ollama','ollama','list')

    def backup(self):
        self.backup_index += 1
        folder=self.state/'backups'/(self.tag+'-'+str(self.backup_index));folder.mkdir(parents=True,exist_ok=True);os.chmod(folder,0o700)
        dest=folder/'postgres.dump'
        args=compose_args(self.root,self.env)+['exec','-T','postgres','pg_dump','-Fc','-U',self.env['POSTGRES_USER'],'-d',self.env['POSTGRES_DB']]
        env=os.environ.copy()
        for key in self.env:env.pop(key,None)
        with dest.open('wb') as f:
            p=subprocess.run(args,stdout=f,stderr=subprocess.PIPE,env=env,timeout=600)
        os.chmod(dest,0o600)
        if p.returncode or dest.stat().st_size<16:raise RuntimeError('Database backup failed: '+self.redact(p.stderr.decode(errors='replace')))
        atomic_write(folder/'postgres.dump.sha256',sha(dest)+'  postgres.dump\n')
        shutil.copy2(self.root/'.env',folder/'env.private');os.chmod(folder/'env.private',0o600)
        self.backup_path=dest
        self.info('Database/config backup: '+str(folder)+'; object files/models are NOT part of this SQL dump')

    def restore_drill(self):
        if not self.backup_path:self.backup()
        temporary='trest_restore_'+secrets.token_hex(8)
        self.psql('CREATE DATABASE "'+temporary+'";\n')
        try:
            args=compose_args(self.root,self.env)+['exec','-T','postgres','pg_restore','--exit-on-error','--no-owner','-U',self.env['POSTGRES_USER'],'-d',temporary]
            env=os.environ.copy()
            for key in self.env:env.pop(key,None)
            with self.backup_path.open('rb') as f:
                p=subprocess.run(args,stdin=f,stdout=subprocess.PIPE,stderr=subprocess.PIPE,env=env,timeout=900)
            if p.returncode:raise RuntimeError('Restore drill failed: '+self.redact(p.stderr.decode(errors='replace')))
            source=self.psql('SELECT version,checksum FROM public.schema_migrations ORDER BY version;\n').stdout
            restored=self.psql('SELECT version,checksum FROM public.schema_migrations ORDER BY version;\n',temporary).stdout
            if source!=restored:raise RuntimeError('Restore drill migration checksum mismatch')
            self.info('Restore into isolated temporary database: PASS; source database unchanged')
        finally:
            assert re.fullmatch(r'trest_restore_[a-f0-9]{16}',temporary)
            self.psql('DROP DATABASE "'+temporary+'" WITH (FORCE);\n')

    def http(self,url,expected_json=None,require_release=False):
        req=urllib.request.Request(url,headers={'User-Agent':'Trest-Installer/'+VERSION})
        with urllib.request.urlopen(req,timeout=15) as r:
            body=r.read(2**20)
            if r.status!=200:raise RuntimeError('HTTP status '+str(r.status))
            if require_release and r.headers.get('X-Trest-Release')!=VERSION:raise RuntimeError('Response belongs to another application, not current Trest edge')
            if expected_json and any(json.loads(body).get(k)!=v for k,v in expected_json.items()):raise RuntimeError('Unexpected health JSON')
        self.info('HTTP PASS: '+url)

    def smoke(self):
        conf=self.config()
        ids=self.dc('ps','-aq',quiet=True).stdout.split()
        if not ids:raise RuntimeError('No containers found')
        containers=json.loads(self.run(['docker','inspect',*ids],quiet=True).stdout)
        by_service={c['Config']['Labels'].get('com.docker.compose.service'):c for c in containers if (c['Config'].get('Labels') or {}).get('com.docker.compose.oneoff','False').lower()!='true'}
        for service in REQUIRED_SERVICES:
            c=by_service.get(service)
            if not c or c['State']['Status']!='running':raise RuntimeError(service+' is not running')
            health=c['State'].get('Health',{}).get('Status')
            if health and health!='healthy':raise RuntimeError(service+' is '+health)
            self.info('Container '+service+': '+(health or 'running (no healthcheck)'))
        # Local HTTP proves local edge routing, not reachability from another ISP.
        base='http://127.0.0.1:'+self.env.get('HTTP_PORT','80')
        if self.env.get('TLS_MODE','off')=='off':
            self.http(base+'/',require_release=True)
            self.http(base+'/ready',{'status':'ready'})
            self.http(base+'/marketplace-api/health',{'status':'ok'})
        elif self.env.get('TLS_MODE')=='auto':
            self.http(self.env['PUBLIC_URL']+'/',require_release=True)
        else:
            self.warn('Public browser trust NOT verified for internal CA mode')
        self.http('http://127.0.0.1:'+self.env.get('WEBUI_PORT','3001')+'/health',{'status':True})
        auth=conf['services']['super-sistema-webui']['environment'].get('WEBUI_AUTH','')
        if str(auth).lower()!='true':raise RuntimeError('WebUI authentication disabled')
        webui_env={a.split('=',1)[0]:a.split('=',1)[1] for a in by_service['super-sistema-webui']['Config']['Env'] if '=' in a}
        if webui_env.get('RESET_CONFIG_ON_START','false').lower()=='true':raise RuntimeError('Unsafe perpetual WebUI config reset remains enabled')
        if not self.a.skip_models:
            script='''import json,urllib.request,os
base="http://ollama:11434/api/"
def post(path,payload):
 req=urllib.request.Request(base+path,data=json.dumps(payload).encode(),headers={"Content-Type":"application/json"})
 with urllib.request.urlopen(req,timeout=240) as r:return json.load(r)
e=post("embed",{"model":os.environ["RAG_EMBEDDING_MODEL"],"input":"test"});assert len(e.get("embeddings",[]))==1 and len(e["embeddings"][0])>0
g=post("generate",{"model":os.environ["DEFAULT_MODELS"],"prompt":"Say OK","stream":False,"options":{"num_predict":4,"num_ctx":1024}});assert g.get("done") is True and g.get("response")
print("Embedding + generation smoke: PASS")
'''
            self.dc('exec','-T','super-sistema-webui','python','-c',script,timeout=300)
        self.warn('External browser/Internet reachability and full business E2E remain separate checks; local HTTP is not external verification')

    def scan_logs(self):
        output=self.dc('logs','--no-color','--since',self.started,'--tail','500',quiet=True).stdout
        hits=[s for s in self.redact(output).splitlines() if re.search(r'panic:|FATAL:|Traceback \(most recent call last\)|Application startup failed|Segmentation fault',s)]
        if hits:raise RuntimeError('Critical current-run log entries:\n'+'\n'.join(hits[-15:]))
        self.info('No matching critical log signatures in this run (not a proof of absence of all bugs)')

    def report(self,code):
        directory=Path(self.a.report_dir) if self.a.report_dir else (SOURCE/'verification/dry-run' if self.a.dry_run else Path('/var/log/trest-systems'))
        directory.mkdir(parents=True,exist_ok=True)
        doc={'version':VERSION,'started_at':self.started,'finished_at':now(),'exit_code':code,'success':code==0,'dry_run':self.a.dry_run,'scope':'package validation' if self.a.dry_run else 'container/HTTP/AI smoke; optional isolated database restore','full_product_e2e':False,'external_browser_verified':False,'install_dir':str(self.root),'public_url':self.env.get('PUBLIC_URL',self.public),'steps':self.steps,'warnings':self.warnings,'errors':self.errors}
        stem=directory/('install-'+self.tag)
        atomic_write(stem.with_suffix('.json'),json.dumps(doc,ensure_ascii=False,indent=2)+'\n')
        text='TREST SYSTEMS '+VERSION+'\n'+ '\n'.join(x['status']+'  '+x['name'] for x in self.steps)+'\n\nОШИБКИ: '+ ('; '.join(self.errors) or 'нет')+'\n\nПРЕДУПРЕЖДЕНИЯ:\n'+'\n'.join(self.warnings)+'\n\nПолная продуктовая E2E-проверка: не выполнена этим smoke-тестом.\nВнешний браузер: отдельно, не проверен с сервера.\n'
        atomic_write(stem.with_suffix('.txt'),text)
        atomic_write(stem.with_suffix('.log'),'\n'.join(self.logs)+'\n')
        self.info('Report: '+str(stem.with_suffix('.txt')))
        if code==0:self.info('Проверки выбранного режима завершены. URL: '+doc['public_url'])

    def execute(self):
        self.stage('Проверка системы',self.system_check)
        if self.a.dry_run:
            self.stage('Целостность полного пакета',self.package_check);return
        if self.a.doctor:
            self.stage('Проверка контейнеров и HTTP',self.smoke);return
        if self.a.repair:
            if not self.env:raise RuntimeError('Existing installation .env not found; repair does not generate new credentials')
            self.stage('Запуск без копирования и изменения конфигурации',lambda:self.start_services(*REQUIRED_SERVICES))
            self.stage('Проверка существующей установки',self.smoke);return
        if self.a.migrate_only:
            self.stage('Миграции',self.migrate);return
        if self.a.backup_only:
            self.stage('Резервная копия БД и конфигурации',self.backup);return
        if self.a.restore_drill_only:
            self.stage('Резервная копия БД',self.backup);self.stage('Изолированное восстановление БД',self.restore_drill);return
        self.stage('Целостность установочного пакета',self.package_check)
        self.stage('Системные компоненты и Docker',self.prerequisites)
        if (self.root/'.env').is_file() and (self.root/'deployments/docker-compose.yml').is_file():
            self.stage('Резервная копия перед изменением существующей установки',self.backup)
            self.stage('Совместимость существующего магазина до изменения файлов',self.check_legacy_shop)
        self.stage('Сохранение секретов и безопасный выбор портов',self.configure)
        self.stage('Копирование проекта без перезаписи .env',self.copy_project)
        self.stage('Конфигурация HTTP/TLS',self.render)
        self.stage('Проверка Compose',lambda:self.dc('config','--quiet'))
        self.stage('Разрешение только HTTP/TLS в существующем firewall',self.configure_firewall)
        self.stage('Импорт/подготовка образов',self.images)
        self.stage('Запуск инфраструктуры',lambda:self.start_services('postgres','redis','minio','ollama'))
        self.stage('Проверка пароля PostgreSQL по TCP',self.password_check)
        self.stage('Резервная копия БД',self.backup)
        self.stage('Проверка совместимости старого магазина',self.check_legacy_shop)
        self.stage('Единый транзакционный мигратор',self.migrate)
        self.stage('Подготовка локальных моделей',self.models)
        self.stage('Запуск всех сервисов и реальные healthchecks',lambda:self.start_services(*REQUIRED_SERVICES))
        self.stage('HTTP и AI smoke-проверки',self.smoke)
        self.stage('Проверка журналов текущего запуска',self.scan_logs)
        if not self.a.no_restore_drill:
            self.stage('Свежая резервная копия БД',self.backup)
            self.stage('Восстановление в отдельную временную БД',self.restore_drill)
        else:self.warn('Isolated database restore drill SKIPPED by flag')


def parser():
    p=argparse.ArgumentParser(description='Trest Systems '+VERSION+' installer (Linux, no production-volume deletion)')
    p.add_argument('--install-dir',default='/opt/trest-systems')
    p.add_argument('--state-dir',default='/var/lib/trest-systems')
    p.add_argument('--report-dir')
    p.add_argument('--domain');p.add_argument('--email');p.add_argument('--public-host')
    p.add_argument('--tls',choices=['auto','off','internal'])
    p.add_argument('--http-port',type=int);p.add_argument('--https-port',type=int)
    p.add_argument('--port-policy',choices=['auto','fail'],default='auto')
    p.add_argument('--wait-timeout',type=int,default=300)
    for f in ('dry-run','doctor','update','repair','offline','rebuild','skip-models','non-interactive','sync-postgres-password','no-restore-drill','migrate-only','backup-only','restore-drill-only'):
        p.add_argument('--'+f,action='store_true')
    p.add_argument('--image-bundle')
    # Existing firewall and SSH rules are never modified without an explicit migration plan.
    p.add_argument('--no-firewall',action='store_true',help='compatibility flag; installer never disables or resets a firewall')
    return p


def main():
    os.umask(0o077)
    a=parser().parse_args()
    try:installer=Installer(a)
    except Exception as e:print('ERROR:',e,file=sys.stderr);return 2
    code=0;lock=None
    try:
        if not a.dry_run:
            path=Path('/var/lock/trest-'+hashlib.sha256(str(installer.root).encode()).hexdigest()[:16]+'.lock')
            lock=path.open('a+')
            try:fcntl.flock(lock.fileno(),fcntl.LOCK_EX|fcntl.LOCK_NB)
            except BlockingIOError:raise RuntimeError('Another installer is using this installation') from None
        installer.execute()
    except (Exception,KeyboardInterrupt) as e:
        code=1;installer.errors.append(installer.redact(str(e) or type(e).__name__));installer.info('[FAIL] '+installer.errors[-1])
    finally:
        installer.report(code)
        if lock:lock.close()
    return code

if __name__=='__main__':sys.exit(main())
