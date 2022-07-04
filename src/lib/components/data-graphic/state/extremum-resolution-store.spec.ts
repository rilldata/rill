import { createExtremumResolutionStore } from "./extremum-resolution-store";
import { get } from "svelte/store";

describe('createExtremumResolutionStore', () => {
  it('instantiates either with undefined or with a concrete value', () => {
    const undefinedStore = createExtremumResolutionStore();
    expect(get(undefinedStore)).toBe(undefined);
    const storeWithPassedValue = createExtremumResolutionStore(10);
    expect(get(storeWithPassedValue)).toBe(10);
  });
  it('instantiates with an empty value if you pass in a single instance value', () => {
    const store = createExtremumResolutionStore(10);
    expect(get(store)).toBe(10);
  });
  it('picks the min value by default if direction not specified. Order should not matter', () => {

    const store1 = createExtremumResolutionStore(10);
    store1.setWithKey('first', 10);
    expect(get(store1)).toBe(10);
    store1.setWithKey('second', 5);
    expect(get(store1)).toBe(5);

    // order should not matter
    const store2 = createExtremumResolutionStore(10);
    store2.setWithKey('second', 5);
    expect(get(store2)).toBe(5);
    store2.setWithKey('first', 10);
    expect(get(store2)).toBe(5);

  });
  it('picks the max value by default if not specified. Order should not matter.', () => {
    const store = createExtremumResolutionStore(10, { direction: 'max' });
    store.setWithKey('first', 10);
    store.setWithKey('second', 5);
    expect(get(store)).toBe(10);

    // order should not matter
    store.setWithKey('first', 10);
    store.setWithKey('second', 5);
    expect(get(store)).toBe(10);
  });

  it('respects an override no matter the extremum values passed in', () => {
    const minStore = createExtremumResolutionStore(10, { direction: 'min' });
    minStore.setWithKey('overriding', 10, true);
    expect(get(minStore)).toBe(10);
    minStore.setWithKey('will not work', 5);
    expect(get(minStore)).toBe(10);
  })

  it('defaults to the next most extreme value when a key is removed', () => {
    const minStore = createExtremumResolutionStore(10, { direction: 'min' });
    minStore.setWithKey('first', 3);
    expect(get(minStore)).toBe(3);
    minStore.setWithKey('second', 2);
    expect(get(minStore)).toBe(2);
    minStore.setWithKey('third', 1);
    expect(get(minStore)).toBe(1);
    minStore.removeKey('third');
    expect(get(minStore)).toBe(2);
    minStore.removeKey('second');
    expect(get(minStore)).toBe(3);
    minStore.removeKey('first');
    expect(get(minStore)).toBe(10);
  })
})