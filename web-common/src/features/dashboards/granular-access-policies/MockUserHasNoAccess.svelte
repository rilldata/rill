<script lang="ts">
  import { selectedMockUserStore } from "./stores";
</script>

<div class="flex-grow flex flex-col justify-center">
  <div class="flex flex-col justify-center gap-y-2 text-center">
    <div class="text-gray-800 text-lg font-semibold">
      This user can't access this dashboard!
    </div>
    <div class="text-gray-500 text-base font-normal">
      If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404
      page.
    </div>
  </div>
</div>
