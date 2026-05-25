import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import {
  runtimeServicePushEnv,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { getContext, setContext } from "svelte";
import { isCloudRuntimeEditEnvironment } from "@rilldata/web-common/features/entity-management/edit-environment.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import {
  parseDotEnv,
  serializeDotEnv,
} from "@rilldata/web-common/features/env-management/dot-env.ts";

const EnvFileStoreKey = "rill:app:env-file-store";

export function createEnvFileStore(runtimeClient: RuntimeClient) {
  const envArtifact = fileArtifacts.getFileArtifact("/.env");
  const envStore = new EnvStore(
    async () => {
      const envBlob = await envArtifact.fetchContent();
      return envBlob ? parseDotEnv(envBlob) : {};
    },
    async (entries) => {
      await runtimeServicePutFile(runtimeClient, {
        path: "/.env",
        blob: serializeDotEnv(entries),
      });
      if (isCloudRuntimeEditEnvironment()) {
        // Only push env on cloud for now. We will revisit this for rill-dev.
        await runtimeServicePushEnv(runtimeClient, {});
      }
    },
  );
  setContext(EnvFileStoreKey, envStore);
  void envStore.pull();
  return eventBus.on("env-file-updated", () => {
    void envStore.pull();
  });
}

export function getEnvFileStore() {
  return getContext<EnvStore>(EnvFileStoreKey);
}
