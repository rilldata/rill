import Axios, { AxiosRequestConfig } from "axios";

let RuntimeUrl = "";
try {
  RuntimeUrl = (window as any).RILL_RUNTIME_URL;
} catch (e) {
  // no-op
}

export const AXIOS_INSTANCE = Axios.create({
  baseURL: RuntimeUrl,
});

export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
