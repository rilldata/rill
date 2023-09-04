/**
 * Run callback once every `time`ms
 */
export function createThrottler(time: number) {
  let timeout: ReturnType<typeof setTimeout>;
  return (callback: () => void) => {
    if (timeout) return;
    timeout = setTimeout(() => {
      console.log(new Date().toISOString(), "sort");
      timeout = undefined;
      callback();
    }, time);
  };
}
