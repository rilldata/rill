import type { AxiosRequestConfig } from "axios";
import Axios from "axios";

export const ADMIN_URL =
  import.meta.env.VITE_RILL_ADMIN_URL || "http://localhost:8080";

export const AXIOS_INSTANCE = Axios.create({
  baseURL: ADMIN_URL,
  withCredentials: true,
});

// TODO: use the new client?
export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
