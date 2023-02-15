import type { AxiosRequestConfig } from "axios";
import Axios from "axios";
import { ADMIN_URL } from "../lib/connection";

export const AXIOS_INSTANCE = Axios.create({
  baseURL: ADMIN_URL,
});

// TODO: use the new client?
export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
