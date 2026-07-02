<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let slots: number;
</script>

<span>
  {slots * 4} GiB RAM, {slots} vCPU
  <span class="text-fg-tertiary text-xs ml-1">
    ({m.status_compute_units({ count: slots })})
  </span>
</span>
