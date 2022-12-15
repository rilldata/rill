<script lang="ts">
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import {
    isDuplicateName,
    renameFileArtifact,
    useAllNames,
  } from "@rilldata/web-local/lib/svelte-query/actions";
  import { runtimeStore } from "../../../application-state-stores/application-store";

  import { useRuntimeServiceRenameFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { WorkspaceHeader } from "..";
  import { notifications } from "../../notifications";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";

  export let modelName: string;
  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: allNamesQuery = useAllNames(runtimeInstanceId);
  const queryClient = useQueryClient();
  const renameModel = useRuntimeServiceRenameFileAndReconcile();

  const switchToModel = async (modelName: string) => {
    if (!modelName) return;

    appStore.setActiveEntity(modelName, EntityType.Model);
  };

  $: switchToModel(modelName);

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Model name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = modelName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, $allNamesQuery.data)) {
      notifications.send({
        message: `Name ${e.target.value} is already in use`,
      });
      e.target.value = modelName; // resets the input
      return;
    }

    try {
      await renameFileArtifact(
        queryClient,
        runtimeInstanceId,
        modelName,
        e.target.value,
        EntityType.Model,
        $renameModel
      );
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  $: titleInput = modelName;
</script>

{#key modelName}
  <WorkspaceContainer assetID={modelName}>
    <div slot="header">
      <WorkspaceHeader
        {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
      />
    </div>
    <div slot="body">
      <ModelBody {modelName} />
    </div>
    <ModelInspector {modelName} slot="inspector" />
  </WorkspaceContainer>
{/key}
