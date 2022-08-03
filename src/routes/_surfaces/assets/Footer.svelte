<script lang="ts">
  import Discord from "$lib/components/icons/Discord.svelte";
  import Docs from "$lib/components/icons/Docs.svelte";

  const lineItems = [
    {
      icon: Docs,
      label: "Docs",
      href: "https://docs.rilldata.com",
      className: "fill-gray-600",
    },
    {
      icon: Discord,
      label: "Feedback",
      href: "http://bit.ly/3jg4IsF",
      className: "fill-gray-500",
    },
  ];
</script>

<div class="flex flex-col bg-gray-50 py-3 gap-y-1">
  {#each lineItems as lineItem}
    <a href={lineItem.href} target="_blank"
      ><div
        class="flex flex-row items-center px-3 py-1 gap-x-2 font-bold hover:bg-gray-200"
      >
        <div>
          <svelte:component
            this={lineItem.icon}
            className={lineItem.className}
            size="1.25em"
          />
        </div>
        {lineItem.label}
      </div></a
    >
  {/each}
</div>
