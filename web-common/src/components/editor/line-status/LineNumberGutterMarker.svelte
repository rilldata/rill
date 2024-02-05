<script lang="ts">
  export let level: "error" | "warning" | "info" | "success" | undefined =
    undefined;
  export let line: number;
  export let active = false;
</script>

<div
  class="grid justify-items-end text-gray-500 text-right pl-1 pr-4"
  class:bg-red-50={level === "error" && !active}
  class:bg-red-100={level === "error" && active}
  class:text-red-600={level === "error" && !active}
  class:text-red-700={level === "error" && active}
  class:bg-yellow-200={level === "warning" && !active}
  class:bg-yellow-300={level === "warning" && active}
  class:text-yellow-500={level === "warning" && !active}
  class:text-yellow-600={level === "warning" && active}
  class:bg-primary-200={level === "info"}
  class:bg-primary-300={level === "info" && !active}
  class:bg-primary-500={level === "info" && active}
  class:text-primary-600={level === "info" && !active}
>
  {line}
</div>
