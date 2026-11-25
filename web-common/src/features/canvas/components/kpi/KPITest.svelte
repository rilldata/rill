<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { ComponentProps } from "svelte";
  import KPI from "./KPI.svelte";

  /**
   * Test-only wrapper that supplies the Tooltip.Provider that the app layout
   * normally provides.
   */

  export let kpiProps: ComponentProps<KPI>;
</script>

<Tooltip.Provider>
  <KPI {...kpiProps} />
</Tooltip.Provider>
