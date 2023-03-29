<script lang="ts">
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { SelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/types-and-utils";

  let selectableItems: SelectableItem[] = [
    { label: "sales", selected: true, visibleInMenu: true },
    { label: "2returns", selected: true, visibleInMenu: true },
    { label: "3new customers", selected: true, visibleInMenu: true },
    { label: "4asdfa", selected: true, visibleInMenu: true },
    { label: "5asdfwet", selected: true, visibleInMenu: true },
    { label: "6qwrtqwer", selected: true, visibleInMenu: true },
    { label: "7salesol;", selected: true, visibleInMenu: true },
    { label: "8returns,n;o;", selected: true, visibleInMenu: true },
    { label: "9new customers6l,", selected: true, visibleInMenu: true },
    { label: "10vbafdg", selected: true, visibleInMenu: true },
    { label: "11wetafasdfwet", selected: true, visibleInMenu: true },
    { label: "12qwrbafbatqwer", selected: true, visibleInMenu: true },
    { label: "13sales", selected: true, visibleInMenu: true },
    { label: "14returagfbvans", selected: true, visibleInMenu: true },
    { label: "15new cwebvaaustomers", selected: true, visibleInMenu: true },
    { label: "16aaeasdfa", selected: true, visibleInMenu: true },
    { label: "17asdfwzxcvawseet", selected: true, visibleInMenu: true },
    { label: "18qwasdfasdrtqwer", selected: true, visibleInMenu: true },
    { label: "19sasdfgalesol;", selected: true, visibleInMenu: true },
    { label: "20retuasdfrns,n;o;", selected: true, visibleInMenu: true },
    { label: "21asdfnew customers6l,", selected: true, visibleInMenu: true },
    { label: "22vbafdfasddg", selected: true, visibleInMenu: true },
    { label: "23wetafasdfasdfwet", selected: true, visibleInMenu: true },
    { label: "24qwrbafbasdfatqwer", selected: true, visibleInMenu: true },
  ];

  let selectableItemsOutput: SelectableItem[];

  $: {
    console.log(
      "selected items",
      selectableItems.filter((x) => x.selected)
    );

    console.log("selectableItemsOutput", selectableItemsOutput);
  }
</script>

<section class="grid grid-flow-row gap-y-4" style="margin-top:40px">
  <h1 class="text-lg">Searchable Filter Menus</h1>
  <SeachableFilterButton
    tooltipText="Choose measures to display"
    {selectableItems}
    bind:selectableItemsOutput
  />
</section>
