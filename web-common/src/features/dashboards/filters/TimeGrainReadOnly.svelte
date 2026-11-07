<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { translateV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  export let timeGrain: V1TimeGrain;
</script>

<Chip type="time" readOnly>
  <svelte:fragment slot="body">
    <div class="text-fg-primary px-2">
      {m.explore_by_grain_prefix()}
      <span class="font-bold">{translateV1TimeGrain(timeGrain)}</span>
    </div>
  </svelte:fragment>
</Chip>
