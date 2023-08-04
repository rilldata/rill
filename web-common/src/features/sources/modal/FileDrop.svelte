<script lang="ts">
  import { goto } from "$app/navigation";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import {
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceUnpackEmpty,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { notifications } from "../../../components/notifications";
  import { appScreen } from "../../../layout/app-store";
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
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceNames = useSourceNames(runtimeInstanceId);
  $: modelNames = useModelNames(runtimeInstanceId);
  $: isProjectInitialized = useIsProjectInitialized(runtimeInstanceId);

  const createSourceMutation = createRuntimeServicePutFileAndReconcile();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  $: createSourceMutationError = ($createSourceMutation?.error as any)?.response
    ?.data;

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

        const sourceError = getSourceError(errors, tableName);
        if ($createSourceMutation.isError || sourceError) {
          // Error
          // Navigate to source page
          goto(`/source/${tableName}`);

          // Telemetry
          emitSourceErrorTelemetry(
            MetricsEventSpace.Workspace,
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
            MetricsEventSpace.Workspace,
            $appScreen,
            BehaviourEventMedium.Drag,
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
