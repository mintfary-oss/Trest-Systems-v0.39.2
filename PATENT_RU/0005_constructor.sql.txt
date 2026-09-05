-- Stage 4: construction object catalog and immutable project versions.
CREATE TABLE IF NOT EXISTS object_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    parameters_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    unit TEXT NOT NULL DEFAULT 'pcs',
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS engineering_systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS finishes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    material_id UUID REFERENCES materials(id) ON DELETE SET NULL,
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE projects ADD COLUMN IF NOT EXISTS object_type_id UUID REFERENCES object_types(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS location JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS parameters JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS architecture_version INTEGER NOT NULL DEFAULT 1 CHECK (architecture_version > 0);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS bim_model_url TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS project_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','review','approved','archived')),
    snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at TIMESTAMPTZ,
    UNIQUE(project_id, version)
);

CREATE TABLE IF NOT EXISTS project_materials (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id),
    quantity NUMERIC(16,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY(project_id, material_id)
);

CREATE TABLE IF NOT EXISTS project_engineering_systems (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    engineering_system_id UUID NOT NULL REFERENCES engineering_systems(id),
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY(project_id, engineering_system_id)
);

CREATE TABLE IF NOT EXISTS project_finishes (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    finish_id UUID NOT NULL REFERENCES finishes(id),
    room TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(16,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY(project_id, finish_id, room)
);

CREATE INDEX IF NOT EXISTS idx_object_types_active ON object_types(active);
CREATE INDEX IF NOT EXISTS idx_materials_category ON materials(category);
CREATE INDEX IF NOT EXISTS idx_engineering_systems_category ON engineering_systems(category);
CREATE INDEX IF NOT EXISTS idx_finishes_category ON finishes(category);
CREATE INDEX IF NOT EXISTS idx_project_versions_project ON project_versions(project_id, version DESC);

INSERT INTO object_types(code,name,description,parameters_schema) VALUES
 ('house','Individual house','Individual residential house','{"area_m2":"number","floors":"integer","rooms":"integer"}'::jsonb),
 ('apartment','Apartment','Apartment unit','{"area_m2":"number","rooms":"integer"}'::jsonb),
 ('commercial','Commercial building','Commercial or public building','{"area_m2":"number","floors":"integer","purpose":"string"}'::jsonb),
 ('industrial','Industrial building','Industrial facility','{"area_m2":"number","height_m":"number","purpose":"string"}'::jsonb)
ON CONFLICT (code) DO NOTHING;

INSERT INTO materials(code,name,category,unit) VALUES
 ('concrete','Concrete','structural','m3'),
 ('rebar','Reinforcement steel','structural','kg'),
 ('brick','Brick','walls','pcs'),
 ('insulation','Thermal insulation','insulation','m2'),
 ('roofing','Roofing material','roof','m2')
ON CONFLICT (code) DO NOTHING;

INSERT INTO engineering_systems(code,name,category) VALUES
 ('electrical','Electrical system','electrical'),
 ('water','Water supply','plumbing'),
 ('sewerage','Sewerage','plumbing'),
 ('heating','Heating','climate'),
 ('ventilation','Ventilation','climate')
ON CONFLICT (code) DO NOTHING;

INSERT INTO finishes(code,name,category) VALUES
 ('paint','Interior paint','walls'),
 ('plaster','Decorative plaster','walls'),
 ('tile','Ceramic tile','floor'),
 ('laminate','Laminate flooring','floor'),
 ('facade','Facade finish','exterior')
ON CONFLICT (code) DO NOTHING;
