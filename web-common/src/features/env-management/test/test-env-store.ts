import { EnvEditSessionVariable } from "@rilldata/web-common/features/env-management/env-edit-session-variable.ts";
import { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils.ts";

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
  initEditValues: Record<string, string> = {},
  initStoreValues: Record<string, string> = {},
) {
  const { testEnvs, envStore } = await makeTestEnvStore(initStoreValues);
  await asyncWait(2); // Wait for time to pass so that edit session has older time
  const envEditSession = new EnvEditSession(envStore);
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

export function envVarsAndNameToObject(
  vars: Map<string, EnvEditSessionVariable>,
) {
  return Object.fromEntries(
    vars.entries().map(([k, e]) => [k, e.mappedEnvVarName]),
  );
}
