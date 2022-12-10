<script lang="ts">
  import {
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceListConnectors,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { runtimeStore } from "../application-state-stores/application-store";
  import { overlay } from "../application-state-stores/overlay-store";
  import { deleteFileArtifact } from "../svelte-query/actions";
  import { useSourceNames } from "../svelte-query/sources";
  import { createSource } from "./navigation/sources/createSource";
  import {
    compileCreateSourceYAML,
    inferSourceName,
  } from "./navigation/sources/sourceUtils";
  import notificationStore from "./notifications/notificationStore";

  const createSourceMutation = useRuntimeServicePutFileAndReconcile();
  const queryClient = useQueryClient();
  const listConnectors = useRuntimeServiceListConnectors();
  const deleteSource = useRuntimeServiceDeleteFileAndReconcile();
  $: sourceNames = useSourceNames($runtimeStore?.instanceId);

  $: connectors = $listConnectors?.data?.connectors;
  let waitingOnSourceImport = false;
  const onRefreshClick = async (connectorType: string, url: string) => {
    const connector = connectors.find(
      (connector) => connector.name === connectorType
    );
    const tableName = inferSourceName(connector, url);
    const yaml = compileCreateSourceYAML(
      { type: connectorType, path: url },
      connector.name
    );
    console.log(tableName, yaml);
    waitingOnSourceImport = true;
    let error;
    try {
      overlay.set({ title: `Importing ${tableName} from ${url}` });
      const errors = await createSource(
        queryClient,
        $runtimeStore.instanceId,
        tableName,
        yaml,
        $createSourceMutation
      );
      error = errors[0];
      if (!error) {
        // no-op
        //dispatch("close");
      } else {
        await deleteFileArtifact(
          queryClient,
          $runtimeStore?.instanceId,
          tableName,
          EntityType.Table,
          $deleteSource,
          $appStore.activeEntity,
          $sourceNames.data,
          false
        );
        notificationStore.send({
          message: error.message,
          type: "error",
        });
      }
    } catch (err) {
      // no-op
    }
    waitingOnSourceImport = false;
    overlay.set(null);
  };

  async function handleCmdV(event: KeyboardEvent) {
    if (document.activeElement !== document.body) {
      console.log("something focused?", document.activeElement);
      return;
    }
    if (event.metaKey && event.key === "v") {
      event.preventDefault();
      console.log("huh");
      const text = await navigator.clipboard.readText();
      console.log("Pasted text: ", text);
      // check if a url
      if (text.startsWith("https://")) {
        // treat as remote file
        await onRefreshClick("https", text);
      } else if (text.startsWith("gs://")) {
        // treat as remote file
        await onRefreshClick("gcs", text);
      } else if (text.startsWith("s3://")) {
        // treat as remote file
        await onRefreshClick("s3", text);
      }
    }
  }
</script>

<svelte:window on:keydown={handleCmdV} />
