<script lang="ts">
  export let pullLeft = false;
  export let pullRight = false;
  export let className = "";
</script>

<div
  style:margin-left={pullLeft ? "-4px" : undefined}
  style:margin-right={pullRight ? "-4px" : undefined}
  class={className}
>
  <slot />
</div>
