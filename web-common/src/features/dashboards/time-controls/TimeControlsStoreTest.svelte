<script lang="ts">
  import type { TimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

  /**
   * This component is used to test the time control store. Without a render the requests are not resolving.
   * TODO: get rid of it once testing is figured out.
   */

  export let timeControlsStore: TimeControlStore;
</script>

<div>
  {$timeControlsStore.timeStart}
</div>
