<script lang="ts">
  export let compact = false;
  $: spacingClasses = compact ? `pl-4 pr-2` : "pl-6 pr-4";
</script>

<header
  style:height={compact ? "40px" : "56px"}
  class="grid items-center text-md {spacingClasses}"
  style:grid-template-columns="auto max-content"
>
  <h1><slot name="title" /></h1>
  <div>
    <slot name="right" />
  </div>
</header>
