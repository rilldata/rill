<script lang="ts">
  import Button from "../Button.svelte";

  export let onClick: (event: any) => void;
  export let primary: boolean = false;
</script>

<Button {onClick} {primary}>
  <slot />
</Button>
