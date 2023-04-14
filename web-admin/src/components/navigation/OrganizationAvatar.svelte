<script lang="ts">
  export let organization: string;
</script>

<div
  style:font-size="9px"
  class="inline-flex items-center justify-center rounded text-white  bg-gray-500"
  style="height: 15px; width: 15px;"
>
  {organization.charAt(0).toUpperCase()}
</div>
