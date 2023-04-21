<script lang="ts">
  import { goto } from "$app/navigation";

  export let statusCode: number;
  export let header: string;
  export let body: string;
</script>

<div class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto">
  <h1
    class="text-8xl font-extrabold bg-gradient-to-b from-[#CBD5E1] to-[#E2E8F0] text-transparent bg-clip-text"
  >
    {statusCode}
  </h1>
  <div class="flex flex-col items-center space-y-2">
    <h2 class="text-lg font-semibold">{header}</h2>
    <p class="text-base text-gray-500 text-center" style="max-width: 530px">
      {body}
    </p>
  </div>
  <button
    class="border border-blue-300 text-blue-600 w-full h-10 hover:bg-slate-100 hover:border-gray-100"
    style="max-width: 400px"
    on:click={() => goto("/")}
    >Back to home
  </button>
  <p class="text-gray-500">
    Need help? Reach out to us on <a href="http://bit.ly/3jg4IsF">Discord</a>
  </p>
</div>
