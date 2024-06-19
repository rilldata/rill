<script lang="ts">
  import Button from "../button/Button.svelte";

  export let variant: "primary" | "secondary" = "primary";
  export let href: string | null = null;
  export let disabled = false;
</script>

<Button type={variant} {href} {disabled} wide on:click><slot /></Button>
