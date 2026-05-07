import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import {
  runtimeServiceGetFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { getContext, setContext } from "svelte";

const EnvFileStoreKey = "rill:app:env-file-store";

export function createEnvFileStore(runtimeClient: RuntimeClient) {
  const envStore = new EnvStore(
    async () => {
      const envBlob = await runtimeServiceGetFile(runtimeClient, {
        path: "/.env",
      });
      const envLines =
        envBlob.blob
          ?.split("\n")
          .map((row) => row.trim())
          .filter((row) => row?.length && !row.match(/^.*#/)) ?? [];
      return Object.fromEntries(
        envLines
          .map((row) => {
            const match = row.match(/^(.*?)=(.*)$/);
            if (!match) return ["", ""];
            const [, key, value] = match;
            return [key, value];
          })
          .filter(([k]) => !!k),
      );
    },
    async (entries) => {
      await runtimeServicePutFile(runtimeClient, {
        path: "/.env",
        blob: Object.values(entries)
          .map(([k, v]) => `${k}=${v}`)
          .join("\n"),
      });
    },
  );
  setContext(EnvFileStoreKey, envStore);
}

export function getEnvFileStore() {
  return getContext<EnvStore>(EnvFileStoreKey);
}
