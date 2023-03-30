<script lang="ts">
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { SelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/types-and-utils";

  let selectableItemsObjs: SelectableItem[] = [
    { label: "sales", selected: true },
    { label: "sales", selected: true },
    { label: "sales", selected: true },
    { label: "2returns", selected: true },
    { label: "3new customers", selected: true },
    { label: "4asdfa", selected: true },
    { label: "5asdfwet", selected: true },
    { label: "6qwrtqwer", selected: true },
    { label: "7salesol;", selected: true },
    { label: "8returns,n;o;", selected: true },
    { label: "9new customers6l,", selected: true },
    { label: "10vbafdg", selected: true },
    { label: "11wetafasdfwet", selected: true },
    { label: "12qwrbafbatqwer", selected: true },
    { label: "13sales", selected: true },
    { label: "14returagfbvans", selected: true },
    { label: "15new cwebvaaustomers", selected: true },
    { label: "16aaeasdfa", selected: true },
    { label: "17asdfwzxcvawseet", selected: true },
    { label: "18qwasdfasdrtqwer", selected: true },
    { label: "19sasdfgalesol;", selected: true },
    { label: "20retuasdfrns,n;o;", selected: true },
    { label: "21asdfnew customers6l,", selected: true },
    { label: "22vbafdfasddg", selected: true },
    { label: "23wetafasdfasdfwet", selected: true },
    { label: "24qwrbafbasdfatqwer", selected: true },
  ];

  let selectableItems = selectableItemsObjs.map((x) => x.label);

  let selectedItems: boolean[] = selectableItems.map((_, i) => i % 2 == 0);

  $: {
    console.log(
      "selected items",
      selectableItems.filter((_, i) => selectedItems[i])
    );
  }
</script>

<section class="grid grid-flow-row gap-y-4" style="margin-top:40px">
  <h1 class="text-lg">Searchable Filter Menus</h1>
  <SeachableFilterButton
    label="Measures"
    tooltipText="Choose measures to display"
    {selectableItems}
    bind:selectedItems
  />
  {#each selectableItems.filter((_, i) => selectedItems[i]) as item}
    {item}&nbsp;
  {/each}
</section>
