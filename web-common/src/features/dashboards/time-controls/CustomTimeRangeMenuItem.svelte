<script lang="ts">
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import MenuItem from "../../../components/menu/core/MenuItem.svelte";

  export let open: boolean;
  export let selected = false;
</script>

<MenuItem animateSelect={false} on:select>
  <div class="w-full flex flex-row flex-grow items-center">
    <div class:font-bold={selected}>Custom range</div>
    <div class="flex-grow" />
    <div class="transition-transform duration-100" class:-rotate-180={open}>
      <CaretDownIcon size="14px" />
    </div>
  </div>
</MenuItem>
