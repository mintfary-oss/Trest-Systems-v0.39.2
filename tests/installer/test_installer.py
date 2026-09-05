import importlib.util,sys,unittest,tempfile,stat,json,subprocess,os
from pathlib import Path
from unittest.mock import patch
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'scripts/installer'))
import auto_install as m
class ConfigTests(unittest.TestCase):
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory();self.root=Path(self.tmp.name);self.p=self.root/'.env'
 def tearDown(self):self.tmp.cleanup()
 def write(self,s):self.p.write_text(s)
 def test_quoted_password_round_trip(self):
  for value in ['a$b#c','abc def',"a'b\\c",'русский пароль','postgresql://u:p@postgres/d?sslmode=disable','']:
   m.update_env(self.p,{'VALUE':value});self.assertEqual(m.read_env(self.p)['VALUE'],value)
 def test_no_shell_execution(self):
  m.update_env(self.p,{'VALUE':'$(touch /tmp/NEVER_EXECUTE_TREST_TEST)'});self.assertEqual(m.read_env(self.p)['VALUE'],'$(touch /tmp/NEVER_EXECUTE_TREST_TEST)')
  self.assertFalse(Path('/tmp/NEVER_EXECUTE_TREST_TEST').exists())
 def test_preserve_unknown_and_comments(self):
  self.write('# comment\nCUSTOM=keep\nPOSTGRES_PASSWORD=original\n');m.update_env(self.p,{'HTTP_PORT':'8099'});s=self.p.read_text();self.assertIn('# comment',s);self.assertIn('CUSTOM=keep',s);self.assertEqual(m.read_env(self.p)['POSTGRES_PASSWORD'],'original')
 def test_duplicate_fails_before_write(self):
  s='A=1\nA=2\n';self.write(s)
  with self.assertRaises(ValueError):m.update_env(self.p,{'A':'3'})
  self.assertEqual(s,self.p.read_text())
 def test_no_syntax_is_silently_ignored(self):
  self.write('INVALID LINE\n')
  with self.assertRaises(ValueError):m.read_env(self.p)
 def test_private_permissions(self):m.update_env(self.p,{'KEY':'abc'});self.assertEqual(stat.S_IMODE(self.p.stat().st_mode),0o600)
 def test_no_multiline_secret(self):
  with self.assertRaises(ValueError):m.update_env(self.p,{'KEY':'a\nb'})
 def test_unquoted_inline_comments(self):self.write('A=foo # comment\n');self.assertEqual(m.read_env(self.p)['A'],'foo')
 def test_quoted_comment_is_data(self):self.write("A='x # y'\n");self.assertEqual(m.read_env(self.p)['A'],'x # y')
 def test_tls_file_not_used_for_http(self):self.assertNotIn(str(self.root/'deployments/tls-ports.yml'),m.compose_args(self.root,{'TLS_MODE':'off'}))
 def test_tls_file_used_for_https(self):self.assertIn(str(self.root/'deployments/tls-ports.yml'),m.compose_args(self.root,{'TLS_MODE':'auto'}))
 def test_project_name_is_explicit(self):a=m.compose_args(self.root,{'COMPOSE_PROJECT_NAME':'unique'});self.assertEqual(a[a.index('--project-name')+1],'unique')
 def test_ownership_not_based_on_container_name(self):
  c={'Name':'trest-fake','Config':{'Labels':{'com.docker.compose.project':'other','com.docker.compose.service':'edge'}},'NetworkSettings':{'Ports':{'80/tcp':[{'HostPort':'80'}]}}}
  self.assertFalse(m.owned_port([c],'trest','edge',80))
 def test_ownership_positive(self):
  c={'Config':{'Labels':{'com.docker.compose.project':'trest','com.docker.compose.service':'edge'}},'NetworkSettings':{'Ports':{'80/tcp':[{'HostPort':'80'}]}}}
  self.assertTrue(m.owned_port([c],'trest','edge',80))
 def installer(self):return m.Installer(m.parser().parse_args(['--install-dir',str(self.root),'--state-dir',str(self.root/'state'),'--report-dir',str(self.root/'reports'),'--public-host','127.0.0.1','--http-port','18493']))
 def test_password_and_port_preserved_repeated_configure(self):
  self.write('POSTGRES_PASSWORD=existing-strong-secret\nCUSTOM=keep\nOLLAMA_PORT=18495\nWEBUI_PORT=18496\nMINIO_PORT=18497\nMINIO_CONSOLE_PORT=18498\n')
  i=self.installer()
  with patch.object(i,'existing_containers',return_value=[]):i.configure();first=m.read_env(self.p);i.configure();second=m.read_env(self.p)
  for k in m.SECRET_KEYS:self.assertEqual(first.get(k),second.get(k),k)
  self.assertEqual(second['POSTGRES_PASSWORD'],'existing-strong-secret');self.assertEqual(second['OLLAMA_PORT'],'18495');self.assertEqual(second['CUSTOM'],'keep')
 def test_existing_db_missing_password_fails(self):
  self.write('COMPOSE_PROJECT_NAME=trest\n');i=self.installer();c={'Config':{'Labels':{'com.docker.compose.project':'trest','com.docker.compose.service':'postgres'}}}
  with patch.object(i,'existing_containers',return_value=[c]),self.assertRaisesRegex(RuntimeError,'password missing'):i.configure()
  self.assertNotIn('POSTGRES_PASSWORD',m.read_env(self.p))
 def test_secret_redaction(self):
  self.write('POSTGRES_PASSWORD=supersecret123\n');i=self.installer();self.assertNotIn('supersecret123',i.redact('value=supersecret123'))
 def test_stage_failure_recorded(self):
  i=self.installer()
  with self.assertRaises(RuntimeError):i.stage('test',lambda:(_ for _ in ()).throw(RuntimeError('bad')))
  self.assertEqual(i.steps[0]['status'],'FAIL');self.assertIn('finished_at',i.steps[0])
 def test_stale_shell_secret_removed(self):
  self.write('POSTGRES_PASSWORD=current-secret\n');i=self.installer()
  with patch.dict(os.environ,{'POSTGRES_PASSWORD':'stale-secret'}),patch('subprocess.run',return_value=subprocess.CompletedProcess([],0,'','')) as run:
   i.run(['echo','ok'],quiet=True)
  self.assertNotIn('POSTGRES_PASSWORD',run.call_args.kwargs['env'])
 def test_doctor_report_not_full_e2e(self):
  i=self.installer();i.report(0);j=json.loads(next((self.root/'reports').glob('*.json')).read_text());self.assertFalse(j['full_product_e2e']);self.assertFalse(j['external_browser_verified'])
 def test_dangerous_install_root_rejected(self):
  with self.assertRaises(ValueError):m.Installer(m.parser().parse_args(['--install-dir','/']))
 def test_sensitive_values_not_sent_in_argv(self):
  source=(ROOT/'scripts/installer/auto_install.py').read_text();self.assertNotIn('-v newpass=',source)
 def test_old_unsafe_cleanups_not_in_active_installer(self):
  for name in ['auto_install.py','apply-migrations.sh','runtime-acceptance.sh']:
   text=(ROOT/'scripts/installer'/name).read_text();self.assertNotIn('down -v',text.replace('no production down -v',''))
if __name__=='__main__':unittest.main(verbosity=2)
