<script lang="ts">
import { createEventDispatcher } from "svelte";
import DataTypeIcon from "$lib/components/data-types/DataTypeIcon.svelte";
import TableHeader from "./TableHeader.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import DataTypeTitle from "$lib/components/tooltip/DataTypeTitle.svelte"
import Pin from "$lib/components/icons/Pin.svelte";

import transientBooleanStore from "$lib/util/transient-boolean-store";
import notificationStore from "$lib/components/notifications/";
import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
import Shortcut from "../tooltip/Shortcut.svelte";
import StackingWord from "../tooltip/StackingWord.svelte";

export let name:string;
export let type:string;
export let pinned = false;

let shiftClicked = transientBooleanStore();
const dispatch = createEventDispatcher();
</script>

<TableHeader {name} {type}>
    <div 
       style:grid-template-columns="210px max-content"
       on:click={async (event) => {
           if (event.shiftKey) {
            await navigator.clipboard.writeText(name);
            notificationStore.send({ message: `copied column name "${name}" to clipboard`});
            // update this to set the active animation in the tooltip text
            shiftClicked.flip();
           }
       }}
       class="
           grid
           items-center
           justify-items-start
           justify-stretch
           select-none
           gap-x-3">

       <Tooltip location="top" alignment="middle" distance={16}>
           <div class="w-full pr-5  flex flex-row gap-x-2 items-center">
               <DataTypeIcon suppressTooltip color={'text-gray-500'} {type} /> 
               <span class="text-ellipsis overflow-hidden whitespace-nowrap ">
                   {name}
               </span>
           </div>
           <TooltipContent slot='tooltip-content'>
               <DataTypeTitle {name} {type} />
               <TooltipShortcutContainer>

                <div>
                    <StackingWord active={$shiftClicked}>
                        copy
                    </StackingWord>
                    column name to clipboard
                </div>
                <Shortcut>
                    <span style='font-family: var(--system);";
                    '>⇧</span> + Click
                </Shortcut>
            </TooltipShortcutContainer>
           </TooltipContent>
       </Tooltip>
       <Tooltip location="top" alignment="middle" distance={16}>
       <button 
           class:text-gray-900={pinned} 
           class:text-gray-400={!pinned}
           class="transition-colors duration-100 justify-self-end"
           on:click={() => {  dispatch("pin" )}}
           >
           <Pin size="16px" />
       </button>
       <TooltipContent slot="tooltip-content">
           {pinned ? 'unpin this column from the right side of the table' : 'pin this column to the right side of the table'}
       </TooltipContent>
       </Tooltip>
   </div>
    </TableHeader>