<script lang="ts">
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { resourceIconMapping } from "../entity-management/resource-icon-mapping";
  import { Settings } from "lucide-svelte";
  import File from "@rilldata/web-common/components/icons/File.svelte";

  export let kind: ResourceKind | undefined;
  export let label: string | undefined;
  export let size = 12;
  export let filePath: string;

  $: icon = kind
    ? resourceIconMapping[kind]
    : filePath === "/.env" || filePath === "/rill.yaml"
      ? Settings
      : File;
</script>

<span class="gap-x-1.5 items-center font-medium flex" title={label}>
  <span class="flex-none">
    <svelte:component this={icon} size="{size}px" />
  </span>
  <p class="truncate">
    {label ?? "Loading..."}
  </p>
</span>
