<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
</script>

<div class="flex flex-col items-center pt-5 pb-10 gap-y-1">
  <Filter size="32px" className="text-slate-300" />
  <div class="flex flex-col items-center">
    <h3 class="text-sm font-semibold text-gray-600">No filters selected</h3>
    <span class="text-sm text-gray-500"
      >To apply filters, close this window and filter your dashboard.</span
    >
  </div>
</div>
