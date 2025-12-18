import { Tabs as TabsPrimitive } from "bits-ui";
import Content from "./TabsContent.svelte";
import List from "./TabsList.svelte";
import Trigger from "./TabsTrigger.svelte";
import UnderlineList from "./UnderlineTabsList.svelte";
import UnderlineTrigger from "./UnderlineTabsTrigger.svelte";

const Root = TabsPrimitive.Root;

export {
  Content,
  List,
  Root,
  //
  Root as Tabs,
  Content as TabsContent,
  List as TabsList,
  Trigger as TabsTrigger,
  Trigger,
  // Underline variant
  UnderlineList as UnderlineTabsList,
  UnderlineTrigger as UnderlineTabsTrigger,
};
