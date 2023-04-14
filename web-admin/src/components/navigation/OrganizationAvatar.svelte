<script lang="ts">
  export let organization: string;
</script>

<div
  class="inline-flex items-center justify-center rounded text-white text-xs font-bold bg-gray-500"
  style="height: 15px; width: 15px;"
>
  {organization.charAt(0).toUpperCase()}
</div>
