<script lang="ts">
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import MenuItem from "../../../components/menu/core/MenuItem.svelte";

  export let open: boolean;
</script>

<MenuItem on:select>
  <div
    class="w-full flex flex-row flex-grow items-center transition-tranform duration-100"
  >
    <div>Custom</div>
    <div class="flex-grow" />
    <div class="transition-transform" class:-rotate-180={open}>
      <CaretDownIcon size="16px" />
    </div>
  </div>
</MenuItem>
