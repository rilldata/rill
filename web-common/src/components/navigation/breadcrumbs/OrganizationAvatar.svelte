<script lang="ts">
  export let organization: string;
</script>

<div
  style:font-size="9px"
  class="inline-flex items-center justify-center rounded-[2px] text-fg-primary bg-surface-elevated"
  style="height: 15px; width: 15px;"
>
  {organization.charAt(0).toUpperCase()}
</div>
