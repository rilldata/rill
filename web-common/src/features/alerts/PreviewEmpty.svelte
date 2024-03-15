<script lang="ts">
  import TableIcon from "../../components/icons/TableIcon.svelte";

  export let topLine: string;
  export let bottomLine: string;
</script>

<div class="pt-5 pb-10 flex flex-col justify-center items-center gap-1">
  <TableIcon size="32px" className="text-slate-300" />
  <div class="flex flex-col justify-center items-center text-sm">
    <div class="text-gray-600 font-semibold">{topLine}</div>
    <div class="text-gray-500 font-normal">
      {bottomLine}
    </div>
  </div>
</div>
