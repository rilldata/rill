<script lang="ts">
  import { goto } from "$app/navigation";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    useValidExplores,
    useValidCanvases,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  // List available explores and canvases
  $: exploresQuery = useValidExplores(instanceId);
  $: canvasesQuery = useValidCanvases(instanceId);
  $: explores = $exploresQuery.data ?? [];
  $: canvases = $canvasesQuery.data ?? [];

  // Build combined dropdown options using file path as value
  $: dashboardOptions = [
    ...explores
      .map((r) => {
        const name = r.meta?.name?.name ?? "";
        const path = r.meta?.filePaths?.[0] ?? "";
        return { value: path, label: name };
      })
      .filter((o) => o.label && o.value),
    ...canvases
      .map((r) => {
        const name = r.meta?.name?.name ?? "";
        const path = r.meta?.filePaths?.[0] ?? "";
        return { value: path, label: name };
      })
      .filter((o) => o.label && o.value),
  ];

  function navigateToDashboard(filePath: string) {
    if (!filePath) return;
    void goto(`/files/${filePath.replace(/^\//, "")}`);
  }
</script>

{#if dashboardOptions.length > 0}
  <div class="flex items-center gap-x-2">
    <span class="text-xs text-fg-secondary">Preview on</span>
    <Select
      id="preview-dashboard"
      label=""
      placeholder="Select dashboard"
      size="sm"
      options={dashboardOptions}
      onChange={navigateToDashboard}
      minWidth={140}
    />
  </div>
{/if}
