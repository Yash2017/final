import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  route?: string;
  http_method?: string;
  data?: object;
  query_params?: string;
  queryText?: string;
  withStreaming?: boolean;
}

export const defaultQuery: Partial<MyQuery> = {
  //constant: 6.5,
  http_method: 'GET',
  route: '',
  withStreaming: false,
  queryText: '',
  query_params: '',
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  username?: string;
  password?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  password?: string;
}
