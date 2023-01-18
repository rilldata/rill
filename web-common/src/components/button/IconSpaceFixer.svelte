<script lang="ts">
  export let pullLeft = false;
  export let pullRight = false;
</script>

<div
  style:margin-left={pullLeft ? "-4px" : undefined}
  style:margin-right={pullRight ? "-4px" : undefined}
>
  <slot />
</div>
