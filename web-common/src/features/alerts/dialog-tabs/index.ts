import { Tabs as TabsPrimitive } from "bits-ui";
import List from "./TabsList.svelte";
import Trigger from "./TabsTrigger.svelte";

const Root = TabsPrimitive.Root;

export {
  List,
  Root,
  //
  Root as Tabs,
  List as TabsList,
  Trigger as TabsTrigger,
  Trigger,
};
