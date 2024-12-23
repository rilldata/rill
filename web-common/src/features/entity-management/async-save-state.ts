import { get, writable } from "svelte/store";
import { fileArtifacts } from "./file-artifacts";

// This class creates a store that manages the state
// of an async, multi-step save operation.
// It is used to show loading/error state in relevant components

// In Rill, saving a file is a multi-step process:
// 1. Initiate the save operation via runtimeServicePutFile
// 2. Receive (asynchronously) a FILE_EVENT_WRITE event from the server
// 3. Re-fetch the file content and check for conflicts

// Saving is not fully "resolved" until step 3 is complete,
// but this happens in a different domain than step 1
// As such, this class uses a deferred promise so that this process can be awaited
// by the various functions that need to wait for the entire save operation to complete

const REJECTION_TIMEOUT = 4000;
const MIN_SAVE_TIME = 500;

export class AsyncSaveState {
  private asyncSavingStore = writable(false);
  private errorStore = writable<null | Error>(null);
  private promise: ReturnType<typeof this.createDeferred> | undefined;
  private touchedStore = writable(false);

  touched = {
    subscribe: this.touchedStore.subscribe,
  };

  touch = (path: string) => {
    const touched = get(this.touched);
    if (touched) return;
    this.touchedStore.set(true);
    fileArtifacts.unsavedFiles.add(path);
  };

  untouch = (path: string) => {
    const touched = get(this.touched);
    if (!touched) return;
    this.touchedStore.set(false);
    fileArtifacts.unsavedFiles.delete(path);
  };

  saving = {
    subscribe: this.asyncSavingStore.subscribe,
  };

  error = {
    subscribe: this.errorStore.subscribe,
  };

  lastSaveTime = 0;

  initiateSave = () => {
    this.lastSaveTime = Date.now();
    this.asyncSavingStore.set(true);

    this.promise = this.createDeferred<void>();
    return this.promise.promise;
  };

  resolve = () => {
    this.promise?.resolve();

    setTimeout(
      () => {
        this.asyncSavingStore.set(false);
        this.errorStore.set(null);
      },
      Math.max(0, MIN_SAVE_TIME - (Date.now() - this.lastSaveTime)),
    );

    this.promise = undefined;
  };

  reject = (e: Error) => {
    this.errorStore.set(e);
    this.asyncSavingStore.set(false);
    this.promise?.reject(e);

    this.promise = undefined;
  };

  private createDeferred<T>(): {
    promise: Promise<T>;
    resolve: (value?: T | PromiseLike<T>) => void;
    reject: (reason?: Error) => void;
  } {
    let resolve!: (value?: T | PromiseLike<T>) => void;
    let reject!: (reason?: Error) => void;

    const promise = new Promise<T>((res, rej) => {
      resolve = res;
      reject = rej;
    });

    const timeoutId = setTimeout(() => {
      reject(new Error("File save timed out."));
    }, REJECTION_TIMEOUT);

    const originalReject = reject;
    reject = (reason?: Error) => {
      clearTimeout(timeoutId);
      originalReject(reason);
    };

    const originalResolve = resolve;
    resolve = (value?: T | PromiseLike<T>) => {
      clearTimeout(timeoutId);
      originalResolve(value);
    };

    return { promise, resolve, reject };
  }
}
