import Axios, { AxiosRequestConfig } from "axios";

export const AXIOS_INSTANCE = Axios.create({
  baseURL: (window as any).RILL_RUNTIME_URL,
});

export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
