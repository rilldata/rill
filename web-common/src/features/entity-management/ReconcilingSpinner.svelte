<script>
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
</script>

<div class="grid place-content-center">
  <div class="grid grid-flow-col">
    <div class="w-6">
      <Spinner
        duration={300 + Math.random() * 200}
        size="1.5em"
        status={EntityStatus.Running}
      />
    </div>
    <div class="w-20">Reconciling...</div>
  </div>
</div>
