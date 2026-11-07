<script lang="ts">
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/personal-files/selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { setQueryClientContext } from "@tanstack/svelte-query";

  /**
   * Test component that lists the viewer's personal canvases the way the
   * project page and the Home menu do.
   */
  let { client }: { client: RuntimeClient } = $props();

  setQueryClientContext(queryClient);
  // The client is fixed for the lifetime of a test.
  // svelte-ignore state_referenced_locally
  const personalCanvases = getPersonalFilteredResources(
    client,
    "acme",
    "ecom",
    ResourceKind.Canvas,
  );
  let resources = $derived($personalCanvases.data ?? []);
</script>

<ul>
  {#each resources as resource (resource.meta?.name?.name)}
    <li>{resource.meta?.name?.name}</li>
  {/each}
</ul>
