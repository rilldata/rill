<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let kind: string;
  export let error: unknown;
</script>

<div
  class="m-auto mt-20 pb-10 flex-col justify-center items-center gap-y-4 inline-flex"
>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-fg-secondary font-semibold">
      {m.resource_error_loading({ kind })}
    </div>
    {#if error instanceof Error}
      <div class="text-fg-muted font-normal">{error.message}</div>
    {/if}
    <div class="text-fg-secondary font-normal">
      {m.resource_error_contact_support()}
    </div>
  </div>
</div>
