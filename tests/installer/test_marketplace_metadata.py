"""Offline schema/model/seed regression tests; no database service is simulated as production."""
import os,sys,unittest,ast
from pathlib import Path
from unittest.mock import patch,MagicMock
import sqlalchemy
from sqlalchemy.schema import CreateTable
from sqlalchemy.dialects import postgresql
ROOT=Path(__file__).resolve().parents[2]
APP=ROOT/'magasin-777/services/evaluator'
sys.path.insert(0,str(APP))
class MarketplaceSchemaTests(unittest.TestCase):
 @classmethod
 def setUpClass(cls):
  with patch.dict(os.environ,{'SECRET_KEY':'metadata-test-only-key','DATABASE_URL':'postgresql://local/test','MARKETPLACE_SCHEMA':'marketplace'}),patch.object(sqlalchemy,'create_engine',return_value=MagicMock()):
   from app.core.database import Base
   from app.models.user import User
   from app.models.order import Order,OrderItem
   from app.models.product import Product,Category,StockMovement
   from app.models.site import Page,SiteSettings,Theme
   cls.base=Base;cls.User=User;cls.Order=Order
 def test_models_isolated_from_core(self):
  for table in self.base.metadata.tables.values():self.assertEqual(table.schema,'marketplace')
 def test_user_primary_key_is_integer_not_core_uuid(self):self.assertIsInstance(self.User.__table__.c.id.type,sqlalchemy.Integer)
 def test_all_foreign_keys_resolve_to_private_schema(self):
  for table in self.base.metadata.tables.values():
   for fk in table.foreign_keys:self.assertEqual(fk.column.table.schema,'marketplace')
 def test_postgres_ddl_compiles(self):
  for table in self.base.metadata.sorted_tables:self.assertIn('marketplace.',str(CreateTable(table).compile(dialect=postgresql.dialect())))
 def test_shop_orders_have_customer_fields(self):self.assertIn('customer_email',self.Order.__table__.c.keys());self.assertNotIn('project_id',self.Order.__table__.c.keys())
 def test_existing_admin_password_not_overwritten(self):
  source=(APP/'app/main.py').read_text();tree=ast.parse(source);node=next(n for n in tree.body if isinstance(n,ast.FunctionDef) and n.name=='_seed_admin');scope={'User':self.User,'DEFAULT_ADMIN_EMAIL':'admin@test.local','DEFAULT_ADMIN_PASSWORD':'new-password','hash_password':MagicMock()};exec(compile(ast.Module(body=[node],type_ignores=[]),'<seed>','exec'),scope)
  db=MagicMock();existing=MagicMock();existing.password_hash='changed-by-user';existing.role='buyer';db.query.return_value.filter.return_value.first.return_value=existing
  scope['_seed_admin'](db)
  self.assertEqual(existing.password_hash,'changed-by-user');self.assertEqual(existing.role,'buyer');scope['hash_password'].assert_not_called()
if __name__=='__main__':unittest.main()
