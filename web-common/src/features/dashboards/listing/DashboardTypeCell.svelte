<script lang="ts">
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;

  $: kind = resource.explore ? ResourceKind.Explore : ResourceKind.Canvas;
</script>

<ResourceTypeBadge {kind} />
