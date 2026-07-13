import { goto } from "$app/navigation";
import { page } from "$app/state";
import { SvelteURL } from "svelte/reactivity";
import {
  ArrayRuneStore,
  type RuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import { tick } from "svelte";

let newParams: [key: string, value: string | null][] = [];

export class UrlParamsState<Val, DefaultVal>
  implements RuneStore<Val, DefaultVal>
{
  public value: Val | DefaultVal;
  private paramValue: string | null = null;

  public constructor(
    private readonly param: string,
    private readonly serializer: (value: Val) => string | null,
    private readonly deserializer: (value: string | null) => Val | DefaultVal,
    defaultVal: Val | DefaultVal,
  ) {
    this.paramValue = page.url.searchParams.get(param);
    this.value = $state(deserializer(this.paramValue) ?? defaultVal);

    $effect(() => {
      const newParamValue = page.url.searchParams.get(this.param);
      if (newParamValue === this.paramValue) return;

      this.paramValue = newParamValue;
      this.value = deserializer(newParamValue);
    });
  }

  public static createStringParam(param: string, defaultValue: string = "") {
    return new UrlParamsState<string, null>(
      param,
      (value) => (value === "" ? null : value),
      (value) => value ?? defaultValue,
      defaultValue,
    );
  }

  public static createStringArrayParam(param: string) {
    return new ArrayRuneStore<string>(
      new UrlParamsState(
        param,
        (value: string[]) => (value.length ? value.join(",") : null),
        (value) => value?.split(",") ?? [],
        [],
      ),
    );
  }

  public getter = () => {
    return this.value;
  };

  public setter = (newValue: Val) => {
    const newParamValue = this.serializer(newValue);
    // Update local state optimistically so in-tick reads see the new value
    // (e.g. two `ArrayRuneStore.toggle` calls in the same tick).
    // The URL write is batched and the constructor's `$effect` reconciles
    // against the actual URL if navigation lands differently.
    this.paramValue = newParamValue;
    this.value = newValue;

    const hasParam = newParams.length > 0;
    newParams.push([this.param, newParamValue]);
    if (!hasParam) void tick().then(flushParams);
  };
}

function flushParams() {
  if (newParams.length === 0) return;

  const newUrl = new SvelteURL(page.url);
  newParams.forEach(([key, value]) => {
    if (value === null) {
      newUrl.searchParams.delete(key);
    } else {
      newUrl.searchParams.set(key, value);
    }
  });

  newParams = [];
  void goto(newUrl, { noScroll: true, keepFocus: true });
}
