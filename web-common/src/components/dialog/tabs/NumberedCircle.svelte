<script lang="ts">
  export let number: number;
  export let bgColor = "";
</script>

<div
  class="w-[18px] h-[18px] text-fg-primary rounded-full inline-flex items-center justify-center {bgColor}"
>
  {number}
</div>
