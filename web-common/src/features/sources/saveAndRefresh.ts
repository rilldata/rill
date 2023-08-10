import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { parse } from "yaml";
import { appScreen } from "../../layout/app-store";
import { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServicePutFileAndReconcile,
} from "../../runtime-client";
import { invalidateAfterReconcile } from "../../runtime-client/invalidation";
import { runtime } from "../../runtime-client/runtime-store";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { fileArtifactsStore } from "../entity-management/file-artifacts-store";
import { EntityType } from "../entity-management/types";
import {
  emitSourceErrorTelemetry,
  emitSourceSuccessTelemetry,
} from "./sourceUtils";

export async function saveAndRefresh(
  queryClient: QueryClient,
  tableName: string,
  yaml: string
) {
  const instanceId = get(runtime).instanceId;
  const filePath = getFilePathFromNameAndType(tableName, EntityType.Table);

  const resp = await runtimeServicePutFileAndReconcile({
    instanceId,
    path: filePath,
    blob: yaml,
    create: false,
    createOnly: false,
    dry: false,
    strict: false,
  });

  invalidateAfterReconcile(queryClient, instanceId, resp);

  // Sometimes, reconcile doesn't detect any changes, but we still need to invalidate the GetFile query
  // One such case is the addition/removal of newlines in the file
  if (resp.affectedPaths.length === 0) {
    queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey(instanceId, filePath)
    );
  }

  // handle errors
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  // emit telemetry
  let yamlObj;
  try {
    yamlObj = parse(yaml);
  } catch (e) {
    // ignore
  }
  if (resp.errors.length > 0) {
    emitSourceErrorTelemetry(
      MetricsEventSpace.RightPanel,
      get(appScreen),
      resp.errors[0]?.message,
      yamlObj?.type || "",
      yamlObj?.uri || ""
    );
  } else {
    emitSourceSuccessTelemetry(
      MetricsEventSpace.Modal,
      get(appScreen),
      BehaviourEventMedium.Button,
      yamlObj?.type || "",
      yamlObj?.uri || ""
    );
  }
}
