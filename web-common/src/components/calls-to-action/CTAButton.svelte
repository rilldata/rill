<script lang="ts">
  import Button from "../button/Button.svelte";

  export let variant: "primary" | "secondary" = "primary";
  export let onClick: ((event: MouseEvent) => void) | undefined = undefined;
  export let href: string | null = null;
  export let rel: string | undefined = undefined;
  export let disabled = false;
  export let gray = false;
  export let submitForm = false;
</script>

<Button
  type={variant}
  {href}
  {rel}
  {disabled}
  wide
  {gray}
  {submitForm}
  {onClick}
>
  <slot />
</Button>
