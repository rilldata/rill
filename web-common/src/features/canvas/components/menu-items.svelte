<script lang="ts" context="module">
  import type { CanvasComponentType } from "./types";
  import ChartIcon from "../icons/ChartIcon.svelte";
  import TableIcon from "../icons/TableIcon.svelte";
  import TextIcon from "../icons/TextIcon.svelte";
  import BigNumberIcon from "../icons/BigNumberIcon.svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

  export type MenuItem = {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  };

  export const menuItems: MenuItem[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon },
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi", label: "KPI", icon: BigNumberIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];
</script>
