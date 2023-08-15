<script lang="ts">
  export let cell: any;
  export let posClass: string = undefined;
  export let negClass: string = undefined;

  let _class = "";
  $: {
    _class = "text-right w-full";
    if (posClass && cell.value > 0) {
      _class += ` ${posClass}`;
    }
    if (negClass && cell.value < 0) {
      _class += ` ${negClass}`;
    }
  }
</script>

<div class={_class}>
  {cell.value.toFixed(2)}
</div>
