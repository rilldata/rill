<script lang="ts">
  import {
    Dialog,
    DialogContent,
  } from "@rilldata/web-common/components/dialog";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";

  /**
   * Test component that opens the edit dialog of an alert the way the alert page does (see `EditAlert`).
   */
  let { alertSpec }: { alertSpec: V1AlertSpec } = $props();
</script>

<!-- The app layout normally supplies the tooltip provider the time controls need. -->
<Tooltip.Provider>
  <Dialog open>
    <DialogContent noClose>
      <AlertFormDataWrapper
        props={{ mode: "edit", alertSpec }}
        onClose={() => {}}
        onCancel={() => {}}
      />
    </DialogContent>
  </Dialog>
</Tooltip.Provider>
