import type { V1Connector } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryServiceQuery } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import { createQuery } from "@tanstack/svelte-query";

export interface OlapInfo {
  replicas: number;
  vcpus: number;
  memory: string;
}

// Standardized columns: replicas, vcpus, memory
const MOTHERDUCK_SQL = `
SELECT
    1 AS replicas,
    MAX(CASE WHEN name = 'threads' THEN CAST(value AS INT) END) AS vcpus,
    MAX(CASE WHEN name = 'memory_limit' THEN value END)         AS memory
FROM duckdb_settings()
WHERE name IN ('threads', 'memory_limit')
`.trim();

// Self-managed ClickHouse
const SELF_MANAGED_CLICKHOUSE_SQL = `
SELECT
    replicas,
    if(cgroup_cpu > 0, cgroup_cpu, os_cpu)                             AS vcpus,
    formatReadableSize(if(server_mem > 0, server_mem, os_mem))         AS memory
FROM (
    SELECT
        (SELECT value FROM system.asynchronous_metrics WHERE metric = 'CGroupMaxCPU')                AS cgroup_cpu,
        (SELECT value FROM system.asynchronous_metrics WHERE metric = 'OSProcessorCount')             AS os_cpu,
        (SELECT toUInt64(value) FROM system.server_settings WHERE name = 'max_server_memory_usage')  AS server_mem,
        (SELECT toUInt64(value) FROM system.asynchronous_metrics WHERE metric = 'OSPhysicalMemoryTotal') AS os_mem
) AS hw
CROSS JOIN (SELECT count() AS replicas FROM system.clusters WHERE cluster = 'default') AS cl
`.trim();

// ClickHouse Cloud — same query for now; update when CHC-specific SQL is confirmed
const CLICKHOUSE_CLOUD_SQL = SELF_MANAGED_CLICKHOUSE_SQL;

export function isMotherDuck(connector: V1Connector): boolean {
  return (
    connector.type === "duckdb" &&
    (String(connector.config?.path ?? "").startsWith("md:") ||
      !!connector.config?.token)
  );
}

export function isClickHouseCloud(connector: V1Connector): boolean {
  if (connector.type !== "clickhouse") return false;
  const cfg = connector.config as Record<string, unknown> | undefined;
  return ["host", "resolved_host", "dsn"].some((field) =>
    String(cfg?.[field] ?? "")
      .toLowerCase()
      .includes(".clickhouse.cloud"),
  );
}

function getOlapInfoSQL(connector: V1Connector | undefined): string | null {
  if (!connector) return null;
  if (isMotherDuck(connector)) return MOTHERDUCK_SQL;
  if (connector.type === "clickhouse") {
    return isClickHouseCloud(connector)
      ? CLICKHOUSE_CLOUD_SQL
      : SELF_MANAGED_CLICKHOUSE_SQL;
  }
  return null;
}

export function useOlapInfo(
  client: RuntimeClient,
  connector: V1Connector | undefined,
) {
  const sql = getOlapInfoSQL(connector);
  const connectorName = connector?.name;

  return createQuery({
    queryKey: ["olap-info", client.instanceId, connectorName],
    queryFn: async ({ signal }) => {
      if (!sql || !connectorName) return null;
      const res = await queryServiceQuery(
        client,
        { connector: connectorName, sql, priority: -1, limit: 0 },
        { signal },
      );
      const row = res.data?.[0];
      if (!row) return null;
      return {
        replicas: Number(row["replicas"] ?? 1),
        vcpus: Number(row["vcpus"] ?? 0),
        memory: String(row["memory"] ?? ""),
      } as OlapInfo;
    },
    enabled: !!sql && !!client.instanceId,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  });
}
