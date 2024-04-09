<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";

  // Button handler: Create new file
  // const createFile = createRuntimeServicePutFileAndReconcile();
  // let path: string;
  // function submit() {
  //   $createFile.mutateAsync({
  //     data: {
  //       instanceId: "default",
  //       path: path,
  //       blob: undefined,
  //       create: true,
  //       createOnly: true,
  //       strict: false,
  //     },
  //   });
  // }
</script>

<div class="p-2">
  <button
    on:click
    class="p-2 bg-primary-50 hover:bg-primary-100 text-primary-700 hover:text-primary-800 w-full flex gap-x-2 items-center font-medium h-7 rounded-sm justify-center"
  >
    <PlusCircleIcon size="14px" />
    <div class="flex gap-x-1 items-center">
      Add
      <CaretDownIcon size="10px" />
    </div>
  </button>
</div>
