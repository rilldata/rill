<script lang="ts">
  import { page } from "$app/stores";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  import { slide } from "svelte/transition";
  import EmbeddedSourceMenuItems from "../EmbeddedSourceMenuItems.svelte";
  import EmbeddedSourceTooltip from "../EmbeddedSourceTooltip.svelte";
  import EndOfPath from "./EndOfPath.svelte";
  export let uri: string;
  export let abbreviatedURI = uri;
  export let connector: string;
  export let cachedSourceName: string;
</script>

<NavigationEntry
  href={`/source/${cachedSourceName}`}
  open={$page.url.pathname === `/source/${cachedSourceName}`}
  name={uri}
  tooltipMaxWidth="300px"
  maxMenuWidth="300px"
>
  <EndOfPath slot="name" path={abbreviatedURI} />

  <svelte:fragment slot="more">
    <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
      <ColumnProfile indentLevel={1} objectName={cachedSourceName} />
    </div>
  </svelte:fragment>
  <svelte:fragment slot="tooltip-content">
    <EmbeddedSourceTooltip sourceName={uri} {connector} />
  </svelte:fragment>

  <EmbeddedSourceMenuItems
    slot="menu-items"
    {cachedSourceName}
    {uri}
    {connector}
  />
</NavigationEntry>
