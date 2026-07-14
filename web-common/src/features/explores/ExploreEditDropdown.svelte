<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let { exploreName }: { exploreName: string } = $props();

  const runtimeClient = useRuntimeClient();

  let exploreQuery = $derived(useExplore(runtimeClient, exploreName));
  let exploreFilePath = $derived(
    $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "",
  );
  let metricsViewFilePath = $derived(
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "",
  );
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button {...props} type="secondary">
        Edit
        <CaretDownIcon />
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
    <DropdownMenu.Item href={getFileHref(exploreFilePath)}>
      <ExploreIcon size="16px" />
      Explore dashboard
    </DropdownMenu.Item>
    <DropdownMenu.Item href={getFileHref(metricsViewFilePath)}>
      <MetricsViewIcon size="16px" />
      Metrics View
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
