<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let isEditForm: boolean;
</script>

<div class="flex flex-col items-center pt-5 pb-10 gap-y-1">
  <Filter size="32px" className="text-gray-300" />
  <div class="flex flex-col items-center">
    <h3 class="text-sm font-semibold text-fg-secondary">{m.alert_no_filters_heading()}</h3>
    <span class="text-sm text-fg-secondary">
      {#if isEditForm}
        {m.alert_no_filters_body()}
      {:else}
        {m.alert_no_filters_hint()}
      {/if}
    </span>
  </div>
</div>
