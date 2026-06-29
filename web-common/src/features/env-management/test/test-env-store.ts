import { EnvEditSessionVariable } from "@rilldata/web-common/features/env-management/env-edit-session-variable.ts";
import { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import type { JSONSchemaObject } from "@rilldata/web-common/features/templates/schemas/types.ts";

export async function makeTestEnvStore(
  initValues: Record<string, string> = {},
) {
  // set and get will be on this map that can be read and written to by tests.
  const testEnvs = initValues;
  const envStore = new EnvStore(
    () => Promise.resolve(initValues),
    async (entries) => {
      Object.keys(testEnvs).forEach((key) => delete testEnvs[key]);
      Object.assign(testEnvs, entries);
    },
  );
  await envStore.pull();

  return {
    testEnvs,
    envStore,
  };
}

export async function makeTestEnvEditSession(
  connectorName: string | undefined,
  schema: JSONSchemaObject | undefined,
  initEditValues: Record<string, string> = {},
  initStoreValues: Record<string, string> = {},
) {
  const { testEnvs, envStore } = await makeTestEnvStore(initStoreValues);
  const envEditSession = new EnvEditSession(
    envStore,
    connectorName ?? "",
    schema,
  );
  Object.entries(initEditValues).forEach(([key, value]) => {
    envEditSession.acquire(key, value);
  });
  return { testEnvs, envStore, envEditSession };
}

export function envMappedVarsAndValuesToObject(
  vars: Map<string, EnvEditSessionVariable>,
) {
  return Object.fromEntries(
    vars.entries().map(([_, e]) => [e.mappedEnvVarName, e.value]),
  );
}
