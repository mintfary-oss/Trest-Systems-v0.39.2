import unittest,yaml,json
from pathlib import Path
R=Path(__file__).resolve().parents[2]
class ReleaseStructure(unittest.TestCase):
 @classmethod
 def setUpClass(c):c.compose=yaml.safe_load((R/'deployments/docker-compose.yml').read_text())
 def test_webui_auth_enabled(self):self.assertEqual(self.compose['services']['super-sistema-webui']['environment']['WEBUI_AUTH'],'true')
 def test_webui_no_perpetual_reset(self):self.assertEqual(self.compose['services']['super-sistema-webui']['environment']['RESET_CONFIG_ON_START'],'false')
 def test_webui_offline_embeddings(self):
  e=self.compose['services']['super-sistema-webui']['environment'];self.assertEqual(e['OFFLINE_MODE'],'true');self.assertEqual(e['HF_HUB_OFFLINE'],'1');self.assertEqual(e['RAG_EMBEDDING_ENGINE'],'ollama')
 def test_internal_services_are_not_public(self):
  for name in ['ollama','minio','super-sistema-webui']:
   for port in self.compose['services'][name].get('ports',[]):self.assertTrue(port.startswith('127.0.0.1:'),(name,port))
  for name in ['api','worker','postgres','redis','marketplace-api','web','nginx']:self.assertFalse(self.compose['services'][name].get('ports'))
 def test_http_only_no_443_publication(self):self.assertEqual(len(self.compose['services']['edge']['ports']),1)
 def test_worker_db_url_required(self):self.assertIn(':?',self.compose['services']['worker']['environment']['DATABASE_URL'])
 def test_restarts_defined(self):
  for s in self.compose['services'].values():self.assertEqual(s['restart'],'unless-stopped')
 def test_migration_cli_is_single_implementation(self):
  s=(R/'scripts/installer/apply-migrations.sh').read_text();self.assertIn('--migrate-only',s);self.assertNotIn('schema_migrations',s)
 def test_no_blind_api_skip(self):self.assertIn('db.CheckMigrations', (R/'cmd/api/main.go').read_text())
 def test_original_migrations_unchanged(self):
  import zipfile
  baseline=R.parent/'Trest-Systems-v0.39.1-FULL-AUTO-INSTALL-PREBUILT.zip'
  if not baseline.exists():self.skipTest('Original baseline ZIP not alongside source')
  with zipfile.ZipFile(baseline) as z:
   for p in sorted((R/'migrations').glob('00[01][0-9]_*.sql')):self.assertEqual(p.read_bytes(),z.read('migrations/'+p.name))
 def test_marketplace_uses_private_schema(self):self.assertEqual(self.compose['services']['marketplace-api']['environment']['MARKETPLACE_SCHEMA'],'marketplace')
 def test_admin_no_reset(self):self.assertNotIn('existing.password_hash =', (R/'magasin-777/services/evaluator/app/main.py').read_text())
 def test_browser_api_has_proxy_prefix(self):self.assertIn('"/marketplace-api"',(R/'magasin-777/frontend/lib/api.js').read_text())
 def test_frontend_node18_removed(self):self.assertNotIn('node:18',(R/'magasin-777/frontend/Dockerfile').read_text())
 def test_minio_real_healthcheck(self):self.assertIn('http://localhost:9000/minio/health/live',self.compose['services']['minio']['healthcheck']['test'])
if __name__=='__main__':unittest.main(verbosity=2)
