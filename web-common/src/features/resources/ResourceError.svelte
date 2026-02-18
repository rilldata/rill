<script lang="ts">
  export let kind: string;
  export let error: unknown;
</script>

<div
  class="m-auto mt-20 pb-10 flex-col justify-center items-center gap-y-4 inline-flex"
>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-fg-secondary font-semibold">Error loading {kind}s</div>
    <div class="text-fg-secondary font-normal">
      If this error persists, please contact support.
    </div>
  </div>
</div>
