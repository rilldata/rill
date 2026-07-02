import { goto } from "$app/navigation";
import { page } from "$app/state";
import { untrack } from "svelte";
import { SvelteURL } from "svelte/reactivity";

export class UrlParamsState<Val, DefaultVal> {
  public value: Val | DefaultVal;
  private paramValue: string | null = null;

  public constructor(
    private readonly param: string,
    private readonly serializer: (value: Val) => string | null,
    deserializer: (value: string | null) => Val | DefaultVal,
    defaultVal: Val | DefaultVal,
  ) {
    this.paramValue = page.url.searchParams.get(param);
    this.value = $state(deserializer(this.paramValue) ?? defaultVal);

    $effect(() => {
      const newParamValue = page.url.searchParams.get(this.param);
      if (newParamValue === this.paramValue) return;

      const newValue = deserializer(newParamValue);
      untrack(() => {
        this.value = newValue;
      });
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

  public getter = () => {
    return this.value;
  };

  public setter = (newValue: Val) => {
    const newParamValue = this.serializer(newValue);
    this.paramValue = newParamValue;
    this.value = newValue;

    const newUrl = new SvelteURL(window.location.href);
    if (newParamValue === null) {
      newUrl.searchParams.delete(this.param);
    } else {
      newUrl.searchParams.set(this.param, newParamValue);
    }
    void goto(newUrl, { noScroll: true, keepFocus: true });
  };
}

export class UrlParamsArrayState<Val> {
  public value: Val[];
  private readonly urlParamState: UrlParamsState<Val[], Val[]>;

  public getter: () => Val[];
  public setter: (newValue: Val[]) => void;

  public constructor(
    private readonly param: string,
    private readonly serializer: (value: Val) => string | null,
    deserializer: (value: string) => Val,
    defaultVal: Val[],
  ) {
    this.urlParamState = new UrlParamsState(
      param,
      (value: Val[]) =>
        value.length ? value.map(this.serializer).join(",") : null,
      (value) => value?.split(",").map(deserializer) ?? defaultVal,
      defaultVal,
    );

    this.value = $derived(this.urlParamState.value);
    this.getter = this.urlParamState.getter;
    this.setter = this.urlParamState.setter;
  }

  public static createStringArrayParam(
    param: string,
    defaultValue: string[] = [],
  ) {
    return new UrlParamsArrayState<string>(
      param,
      (value) => value,
      (value) => value,
      defaultValue,
    );
  }

  public toggle = (newValue: Val) => {
    const newTags = this.value.includes(newValue)
      ? this.value.filter((v) => v !== newValue)
      : [...this.value, newValue];
    this.setter(newTags);
  };
}
