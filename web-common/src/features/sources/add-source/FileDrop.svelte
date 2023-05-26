<script lang="ts">
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import {
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceUnpackEmpty,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModelNames } from "../../models/selectors";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { useIsProjectInitialized } from "../../welcome/is-project-initialized";
  import {
    compileCreateSourceYAML,
    sourceErrorTelemetryHandler,
  } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { uploadTableFiles } from "./file-upload";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { SourceConnectionType } from "../../../metrics/service/SourceEventTypes";

  export let showDropOverlay: boolean;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const createSourceMutation = createRuntimeServicePutFileAndReconcile();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const uploadedFiles = uploadTableFiles(
      Array.from(e?.dataTransfer?.files),
      [$sourceNames?.data, $modelNames?.data],
      $runtime.instanceId
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
        const errors = await createSource(
          queryClient,
          runtimeInstanceId,
          tableName,
          yaml,
          $createSourceMutation
        );

        if (errors) {
          sourceErrorTelemetryHandler(
            MetricsEventSpace.Workspace,
            MetricsEventScreenName.Source,
            errors,
            SourceConnectionType.Local,
            filePath
          );
        }
      } catch (err) {
        console.error(err);
      }
    }
  };
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    class="w-screen h-screen grid place-content-center"
    on:dragenter|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:drag|preventDefault|stopPropagation
    on:drop|preventDefault|stopPropagation={handleSourceDrop}
    on:mouseup|preventDefault|stopPropagation={() => {
      showDropOverlay = false;
    }}
  >
    <div
      class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all text-3xl"
    >
      <span class="place-content-center">
        drop your files to add new source
      </span>
    </div>
  </div>
</Overlay>
