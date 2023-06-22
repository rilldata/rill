SELECT
    runtime_host,
    SUM(slots)
FROM deployments
GROUP BY runtime_host;
