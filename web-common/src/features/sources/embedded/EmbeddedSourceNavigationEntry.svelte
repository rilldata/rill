<script lang="ts">
  import { page } from "$app/stores";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import { slide } from "svelte/transition";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import EmbeddedSourceEntry from "./EmbeddedSourceEntry.svelte";
  import EmbeddedSourceMenuItems from "./EmbeddedSourceMenuItems.svelte";
  import EmbeddedSourceTooltip from "./EmbeddedSourceTooltip.svelte";

  export let path: string;
  export let connector: string;
  export let cachedSourceName: string;
</script>

<NavigationEntry
  href={`/source/${cachedSourceName}`}
  open={$page.url.pathname === `/source/${cachedSourceName}`}
  name={path}
  tooltipMaxWidth="300px"
  maxMenuWidth="300px"
>
  <EmbeddedSourceEntry slot="name" {connector} {path} />

  <svelte:fragment slot="more">
    <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
      <ColumnProfile indentLevel={1} objectName={cachedSourceName} />
    </div>
  </svelte:fragment>
  <svelte:fragment slot="tooltip-content">
    <EmbeddedSourceTooltip sourceName={path} {connector} />
  </svelte:fragment>

  <EmbeddedSourceMenuItems
    slot="menu-items"
    {cachedSourceName}
    uri={path}
    {connector}
  />
</NavigationEntry>
