-- Seed script for stress-testing the user group management UI.
-- Creates 25 fake users and 25 fake projects in the target org.
--
-- Seed:
--   psql postgres://postgres:postgres@localhost:5432/postgres \
--     -v org=rilldata -f scripts/seed-stress-test.sql
--
-- Cleanup:
--   psql postgres://postgres:postgres@localhost:5432/postgres \
--     -v org=rilldata -v cleanup=true -f scripts/seed-stress-test.sql

\if :{?org}
\else
  \set org 'rilldata'
\endif

\if :{?cleanup}
  SET app.seed_cleanup = :'cleanup';
\else
  SET app.seed_cleanup = 'false';
\endif

SET app.seed_org = :'org';

DO $$
DECLARE
  v_org_name TEXT := current_setting('app.seed_org');
  v_cleanup  TEXT := current_setting('app.seed_cleanup');
  v_org_id   UUID;
  v_role_id  UUID;
  v_user_id  UUID;
  v_emails   TEXT[] := ARRAY[
    'alice.johnson@example.com',  'bob.chen@example.com',      'carol.martinez@example.com',
    'david.lee@example.com',      'emma.wilson@example.com',   'frank.garcia@example.com',
    'grace.kim@example.com',      'henry.patel@example.com',   'iris.rodriguez@example.com',
    'jack.thompson@example.com',  'kate.brown@example.com',    'liam.davis@example.com',
    'maya.anderson@example.com',  'noah.white@example.com',    'olivia.harris@example.com',
    'paul.jackson@example.com',   'quinn.taylor@example.com',  'rachel.moore@example.com',
    'sam.nguyen@example.com',     'tina.clark@example.com',    'uma.scott@example.com',
    'victor.lewis@example.com',   'wendy.hall@example.com',    'xavier.young@example.com',
    'yara.walker@example.com'
  ];
  v_names    TEXT[] := ARRAY[
    'Alice Johnson',   'Bob Chen',        'Carol Martinez',
    'David Lee',       'Emma Wilson',     'Frank Garcia',
    'Grace Kim',       'Henry Patel',     'Iris Rodriguez',
    'Jack Thompson',   'Kate Brown',      'Liam Davis',
    'Maya Anderson',   'Noah White',      'Olivia Harris',
    'Paul Jackson',    'Quinn Taylor',    'Rachel Moore',
    'Sam Nguyen',      'Tina Clark',      'Uma Scott',
    'Victor Lewis',    'Wendy Hall',      'Xavier Young',
    'Yara Walker'
  ];
  v_projects TEXT[] := ARRAY[
    'analytics-dashboard',  'sales-pipeline',       'marketing-metrics',
    'product-analytics',    'revenue-tracking',     'user-behavior',
    'customer-insights',    'growth-metrics',       'inventory-analysis',
    'financial-reporting',  'ops-dashboard',        'support-metrics',
    'data-quality',         'campaign-performance', 'churn-analysis',
    'acquisition-funnel',   'retention-metrics',    'cohort-analysis',
    'ab-testing',           'ml-predictions',       'event-tracking',
    'session-analytics',    'geo-distribution',     'pricing-analysis',
    'engineering-metrics'
  ];
BEGIN
  SELECT id INTO v_org_id FROM orgs WHERE lower(name) = lower(v_org_name);
  IF v_org_id IS NULL THEN
    RAISE EXCEPTION 'Org "%" not found', v_org_name;
  END IF;

  -- Cleanup mode: remove seeded data and exit
  IF v_cleanup = 'true' THEN
    DELETE FROM users WHERE email = ANY(v_emails);
    DELETE FROM projects WHERE org_id = v_org_id AND name = ANY(v_projects);
    RAISE NOTICE 'Cleaned up seeded users and projects from org "%"', v_org_name;
    RETURN;
  END IF;

  SELECT id INTO v_role_id FROM org_roles WHERE lower(name) = 'viewer';

  -- Create users and add them to the org as viewers
  FOR i IN 1..array_length(v_emails, 1) LOOP
    INSERT INTO users (email, display_name)
    VALUES (v_emails[i], v_names[i])
    ON CONFLICT (lower(email)) DO NOTHING;

    SELECT id INTO v_user_id FROM users WHERE lower(email) = lower(v_emails[i]);

    INSERT INTO users_orgs_roles (user_id, org_id, org_role_id)
    VALUES (v_user_id, v_org_id, v_role_id)
    ON CONFLICT (user_id, org_id) DO NOTHING;
  END LOOP;

  -- Create stub projects
  FOR i IN 1..array_length(v_projects, 1) LOOP
    INSERT INTO projects (org_id, name, description, public, region, prod_branch, prod_olap_driver, prod_olap_dsn, prod_slots)
    VALUES (v_org_id, v_projects[i], '', false, '', 'main', 'duckdb', '', 2)
    ON CONFLICT DO NOTHING;
  END LOOP;

  RAISE NOTICE 'Seeded % users and % projects into org "%"',
    array_length(v_emails, 1), array_length(v_projects, 1), v_org_name;
END $$;
