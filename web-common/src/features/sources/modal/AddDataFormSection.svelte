<script lang="ts">
  export let id: string;
  export let enhance: any;
  export let onSubmit: (e?: Event) => void;
</script>

<form
  {id}
  class="pb-5 flex-grow overflow-y-auto"
  use:enhance
  onsubmit={(e) => {
    e.preventDefault();
    onSubmit(e);
  }}
>
  <slot />
</form>
