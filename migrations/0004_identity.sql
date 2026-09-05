-- Stage 3: organizations, membership, permissions and verification.
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('customer','contractor','supplier')),
    legal_name TEXT NOT NULL,
    registration_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending','verified','rejected')),
    rating NUMERIC(4,2) NOT NULL DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
    geography JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    membership_role TEXT NOT NULL CHECK (membership_role IN ('owner','admin','member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id,user_id)
);
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS role_permissions (
    role TEXT NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);
INSERT INTO permissions(code,description) VALUES
 ('profile.read','Read own profile'),('organization.create','Create organization'),
 ('organization.read','Read organization'),('organization.manage','Manage organization'),
 ('organization.verify','Verify organization'),('project.create','Create project'),
 ('order.create','Create order'),('admin.audit','Access administrative audit data')
ON CONFLICT (code) DO NOTHING;
INSERT INTO role_permissions(role,permission_id)
SELECT 'customer',id FROM permissions WHERE code IN ('profile.read','organization.create','organization.read','organization.manage','project.create','order.create')
ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role,permission_id)
SELECT 'contractor',id FROM permissions WHERE code IN ('profile.read','organization.create','organization.read','organization.manage','project.create')
ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role,permission_id)
SELECT 'supplier',id FROM permissions WHERE code IN ('profile.read','organization.create','organization.read','organization.manage')
ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role,permission_id)
SELECT 'admin',id FROM permissions
ON CONFLICT DO NOTHING;
CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_status ON organizations(verification_status);
