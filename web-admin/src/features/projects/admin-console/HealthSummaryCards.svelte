<script lang="ts">
  export let totalProjects: number;
  export let healthyCount: number;
  export let errorCount: number;
</script>

<div class="grid grid-cols-3 gap-4">
  <div class="rounded-lg border border-gray-200 p-4">
    <p class="text-sm text-gray-500">Total Projects</p>
    <p class="text-3xl font-semibold text-gray-900">{totalProjects}</p>
  </div>
  <div class="rounded-lg border border-green-200 bg-green-50 p-4">
    <p class="text-sm text-green-600">Healthy</p>
    <p class="text-3xl font-semibold text-green-700">{healthyCount}</p>
  </div>
  <div class="rounded-lg border border-red-200 bg-red-50 p-4">
    <p class="text-sm text-red-600">Erroring</p>
    <p class="text-3xl font-semibold text-red-700">{errorCount}</p>
  </div>
</div>
