<script lang="ts">
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { useIsModelingSupportedForCurrentOlapDriver } from "@rilldata/web-common/features/tables/selectors";
  import GoToDashboardButton from "@rilldata/web-common/features/metrics-views/workspace/GoToDashboardButton.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { goto } from "$app/navigation";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/workspace/editor/MetricsEditor.svelte";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/workspace/inspector/MetricsInspector.svelte";

  export let data;

  $: metricViewName = data.file.name;
  $: filePath = data.file.path;

  const { readOnly } = featureFlags;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: showInspector = $isModelingSupportedForCurrentOlapDriver.data;
  $: initLocalUserPreferenceStore(metricViewName);

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
    );
    if (newRoute) await goto(newRoute);
  };
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<WorkspaceContainer inspector={showInspector}>
  <WorkspaceHeader
    slot="header"
    on:change={onChangeCallback}
    showInspectorToggle={showInspector}
    titleInput={metricViewName}
  >
    <GoToDashboardButton {filePath} slot="cta" />
  </WorkspaceHeader>
  <MetricsEditor {filePath} slot="body" />
  <MetricsInspector {filePath} slot="inspector" />
</WorkspaceContainer>
