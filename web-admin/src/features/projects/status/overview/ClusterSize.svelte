<script lang="ts">
  export let slots: number;
</script>

<span>
  {slots * 4} GiB RAM, {slots} vCPU
  <span class="text-fg-tertiary text-xs ml-1">
    ({slots}
    {slots === 1 ? "Compute unit" : "Compute units"})
  </span>
</span>
