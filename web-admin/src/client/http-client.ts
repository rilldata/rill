import type { AxiosRequestConfig } from "axios";
import Axios from "axios";

export const AXIOS_INSTANCE = Axios.create({
  baseURL: "http://localhost:8080",
});

// TODO: use the new client?
export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
