import "./wasm_exec.js";

const go = new Go();
WebAssembly.instantiateStreaming(fetch("rill-time.wasm"), go.importObject).then(
  (result) => {
    go.run(result.instance);
  },
);

export async function parseRillTimeSyntax(time: string) {
  return window.parseRillTime(time);
}
