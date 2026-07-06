import { goto } from "$app/navigation";
import { page } from "$app/state";
import { SvelteURL } from "svelte/reactivity";
import {
  ArrayRuneStore,
  type RuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

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
    const newUrl = new SvelteURL(window.location.href);
    if (newParamValue === null) {
      newUrl.searchParams.delete(this.param);
    } else {
      newUrl.searchParams.set(this.param, newParamValue);
    }
    void goto(newUrl, { noScroll: true, keepFocus: true });
  };
}
