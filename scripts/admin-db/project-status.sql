SELECT
    o.name AS org_name,
    p.name AS project_name,
    d.status,
    d.updated_on,
    d.logs
FROM deployments d
JOIN projects p ON d.project_id=p.id
JOIN orgs o on p.org_id=o.id;
