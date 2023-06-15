<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Callout } from "@rilldata/web-common/components/callout";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/add-source/file-upload";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen, appStore } from "@rilldata/web-common/layout/app-store";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    createRuntimeServiceDeleteFileAndReconcile,
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceUnpackEmpty,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { useModelNames } from "../../models/selectors";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import {
    compileCreateSourceYAML,
    getSourceError,
    emitSourceErrorTelemetry,
    emitSourceSuccessTelemetry,
  } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { hasDuckDBUnicodeError, niceDuckdbUnicodeError } from "./errors";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { SourceConnectionType } from "../../../metrics/service/SourceEventTypes";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const createSourceMutation = createRuntimeServicePutFileAndReconcile();
  const deleteSource = createRuntimeServiceDeleteFileAndReconcile();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  $: createSourceMutationError = ($createSourceMutation?.error as any)?.response
    ?.data;

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  const handleDeleteSource = async (tableName: string) => {
    await deleteFileArtifact(
      queryClient,
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $deleteSource,
      $appStore.activeEntity,
      $sourceNames.data,
      false
    );
  };

  let errors;

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(
      files,
      [$sourceNames?.data, $modelNames?.data],
      $runtime.instanceId,
      false
    );
    for await (const { tableName, filePath } of uploadedFiles) {
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
      if (!errors?.length) {
        dispatch("close");
      } else {
        // if the upload didn't work, delete the source file.
        handleDeleteSource(tableName);
      }

      const sourceError = getSourceError(errors, tableName);
      if ($createSourceMutation.isError || sourceError) {
        emitSourceErrorTelemetry(
          MetricsEventSpace.Modal,
          $appScreen,
          createSourceMutationError?.message ?? sourceError?.message,
          SourceConnectionType.Local,
          filePath
        );
      } else {
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
  {#if errors?.length}
    <div transition:slide={{ duration: LIST_SLIDE_DURATION * 2 }}>
      <Callout level="error">
        <ul style:max-width="400px">
          {#each errors as error}
            <li>
              {hasDuckDBUnicodeError(error.message)
                ? niceDuckdbUnicodeError(error.message)
                : error.message}
            </li>
          {/each}
        </ul>
      </Callout>
    </div>
  {/if}
</div>
