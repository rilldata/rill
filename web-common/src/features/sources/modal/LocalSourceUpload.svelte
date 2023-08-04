<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceUnpackEmpty,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { notifications } from "../../../components/notifications";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { SourceConnectionType } from "../../../metrics/service/SourceEventTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModelNames } from "../../models/selectors";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import { createModelFromSourceV2 } from "../createModel";
  import {
    compileCreateSourceYAML,
    emitSourceErrorTelemetry,
    emitSourceSuccessTelemetry,
    getSourceError,
  } from "../sourceUtils";
  import { createSource } from "./createSource";

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const createSourceMutation = createRuntimeServicePutFileAndReconcile();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  $: createSourceMutationError = ($createSourceMutation?.error as any)?.response
    ?.data;

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$sourceNames?.data, $modelNames?.data],
      $runtime.instanceId,
      false
    );
    for await (const { tableName, filePath } of uploadedFiles) {
      let errors;

      try {
        // If project is uninitialized, initialize an empty project
        if (!$isProjectInitialized.data) {
          $unpackEmptyProject.mutate({
            instanceId: $runtime.instanceId,
            data: {
              title: EMPTY_PROJECT_TITLE,
            },
          });
        }

        const yaml = compileCreateSourceYAML(
          {
            sourceName: tableName,
            path: filePath,
          },
          "local_file"
        );

        // TODO: errors
        errors = await createSource(
          queryClient,
          runtimeInstanceId,
          tableName,
          yaml,
          $createSourceMutation
        );
      } catch (err) {
        // no-op
      }

      overlay.set(null);
      dispatch("close");

      const sourceError = getSourceError(errors, tableName);
      if ($createSourceMutation.isError || sourceError) {
        // Error
        // Navigate to source page
        goto(`/source/${tableName}`);

        // Telemetry
        emitSourceErrorTelemetry(
          MetricsEventSpace.Modal,
          $appScreen,
          createSourceMutationError?.message ?? sourceError?.message,
          SourceConnectionType.Local,
          filePath
        );
      } else {
        // Success
        // Create a `select *` model
        const newModelName = await createModelFromSourceV2(
          queryClient,
          tableName
        );

        // Navigate to new model
        goto(`/model/${newModelName}?focus`);

        // Show toast message
        notifications.send({
          message: `Data source imported. Start modeling it here.`,
        });

        // Telemetry
        emitSourceSuccessTelemetry(
          MetricsEventSpace.Modal,
          $appScreen,
          BehaviourEventMedium.Button,
          SourceConnectionType.Local,
          filePath
        );
      }
    }
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV, JSON or Parquet file
  </Button>
</div>
