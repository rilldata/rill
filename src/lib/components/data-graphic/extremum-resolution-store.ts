/**
 * extremum-resolution-store
 * -------------------------
 * This specialized store handles the resolution of plot bounds based
 * on multiple extrema. If multiple components set a maximum value within
 * the same namespace, the store will automatically pick the largest value
 * for the store value. This enables us to determine, for instance, if
 * multiple lines are on the same chart, which ones determine the bounds.
 */
import { cubicOut } from "svelte/easing";
import { writable, derived } from "svelte/store";
import { tweened } from "svelte/motion";
import { min, max } from 'd3-array'

const LINEAR_SCALE_STORE_DEFAULTS = {
    duration: 0,
    easing: cubicOut,
    direction: 'min',
    namespace: undefined
}

interface Options {
    duration:number,
    easing:Function
}

interface extremumArgs {
    duration?:number,
    easing?:(t:number) => number,
    direction?:string,
}

interface Extremum {
    value:number,
    override?:boolean
}

const extremaFunctions = { min, max }

export function createExtremumResolutionStore(initialValue, passedArgs:extremumArgs = {}) {
    let args = {...LINEAR_SCALE_STORE_DEFAULTS, ...passedArgs};
    let storedValues = writable({});
    let valueTween = tweened(initialValue, { duration: args.duration, easing: args.easing });

    /**
     * 
     * @param key 
     * @param value 
     * @param override 
     */
    function _update(key:string, value:(number|Date), override = false) {
        storedValues.update((storeValue) => {
            if (!(key in storeValue)) storeValue[key] = {value: undefined, override: false}
            storeValue[key].value = value;
            storeValue[key].override = override;
            return storeValue;
        })
    };

    function _remove(key) {
        storedValues.update((storeValue) => {
            delete storeValue[key];
            return storeValue;
        })
    }

    const domainExtents = derived(storedValues, ($storedValues) => {
        let extremum;
        Object.values($storedValues).forEach((entry:Extremum) => {
            extremum = entry.override && entry.value !== undefined ? entry : 
                extremaFunctions[args.direction]([extremum, entry.value]);
        })
        return extremum;
    }, undefined);

    // set the final tween with the value.
    domainExtents.subscribe(value => {
        if (value !== undefined) {
            valueTween.set(value);
        }
    })

    const returnedStore = {
        subscribe: valueTween.subscribe,
        setWithKey(key, value = undefined, override = undefined) {
            _update(key, value, override);
            
        },
        removeKey(key:string) {
            _remove(key);
        }
    };
    return returnedStore;
}