<script lang="ts">
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import MenuItem from "../../../components/menu/core/MenuItem.svelte";
  import { TimeRangeName } from "./time-control-types";

  export let open: boolean;
</script>

<MenuItem animateSelect={false} on:select>
  <div class="w-full flex flex-row flex-grow items-center">
    <div>{TimeRangeName.Custom}</div>
    <div class="flex-grow" />
    <div class="transition-transform duration-100" class:-rotate-180={open}>
      <CaretDownIcon size="16px" />
    </div>
  </div>
</MenuItem>
