// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

import * as url from "url";
import { Configuration } from "./configuration";

const BASE_PATH = "/api".replace(/\/+$/, "");

export interface FetchAPI {
    (url: string, init?: any): Promise<Response>;
}

export interface FetchArgs {
    url: string;
    options: any;
}

export class BaseAPI {
    protected configuration?: Configuration;

    constructor(configuration?: Configuration, protected basePath: string = BASE_PATH, protected fetch: FetchAPI = window.fetch) {
        if (configuration) {
            this.configuration = configuration;
            this.basePath = configuration.basePath || this.basePath;
        }
    }
}

export class RequiredError extends Error {
    name = "RequiredError";
    constructor(public field: string, msg?: string) {
        super(msg);
    }
}

export interface GetSerialsReply {
    serials: Serials;
}

export interface ImportSchedulesReply {
    error?: string;
}

export interface ImportSchedulesRequest {
    schedules: Schedules;
}

export type Level = number;

export type Levels = Array<Level>;

export interface Schedule {
    start: Time;
    stop: Time;
    levels: Levels;
    serials: Serials;
}

export type Schedules = Array<Schedule>;

export type Serial = number;

export type Serials = Array<Serial>;

export type Time = number;

export const DefaultApiFetchParamCreator = function (configuration?: Configuration) {
    return {
        apiGetSerials(options: any = {}): FetchArgs {
            const localVarPath = `/get-serials`;
            const localVarUrlObj = url.parse(localVarPath, true);
            const localVarRequestOptions = Object.assign({ method: 'GET' }, options);
            const localVarHeaderParameter = {} as any;
            const localVarQueryParameter = {} as any;

            localVarUrlObj.query = Object.assign({}, localVarUrlObj.query, localVarQueryParameter, options.query);
            // fix override query string Detail: https://stackoverflow.com/a/7517673/1077943
            delete localVarUrlObj.search;
            localVarRequestOptions.headers = Object.assign({}, localVarHeaderParameter, options.headers);

            return {
                url: url.format(localVarUrlObj),
                options: localVarRequestOptions,
            };
        },

        apiImportSchedules(body: ImportSchedulesRequest, options: any = {}): FetchArgs {
            // verify required parameter 'body' is not null or undefined
            if (body === null || body === undefined) {
                throw new RequiredError('body', 'Required parameter body was null or undefined when calling apiImportSchedules.');
            }
            const localVarPath = `/import-schedules`;
            const localVarUrlObj = url.parse(localVarPath, true);
            const localVarRequestOptions = Object.assign({ method: 'POST' }, options);
            const localVarHeaderParameter = {} as any;
            const localVarQueryParameter = {} as any;

            localVarHeaderParameter['Content-Type'] = 'application/json';

            localVarUrlObj.query = Object.assign({}, localVarUrlObj.query, localVarQueryParameter, options.query);
            // fix override query string Detail: https://stackoverflow.com/a/7517673/1077943
            delete localVarUrlObj.search;
            localVarRequestOptions.headers = Object.assign({}, localVarHeaderParameter, options.headers);
            const needsSerialization = localVarRequestOptions.headers['Content-Type'] === 'application/json';
            localVarRequestOptions.body = needsSerialization ? JSON.stringify(body || {}) : (body || "");

            return {
                url: url.format(localVarUrlObj),
                options: localVarRequestOptions,
            };
        },
    }
};

export const DefaultApiFp = function (configuration?: Configuration) {
    return {
        apiGetSerials(options?: any): (fetch?: FetchAPI, basePath?: string) => Promise<GetSerialsReply> {
            const localVarFetchArgs = DefaultApiFetchParamCreator(configuration).apiGetSerials(options);
            return (fetch: FetchAPI = window.fetch, basePath: string = BASE_PATH) => {
                return fetch(basePath + localVarFetchArgs.url, localVarFetchArgs.options).then((response) => {
                    if (response.status >= 200 && response.status < 300) {
                        return response.json();
                    } else {
                        throw response;
                    }
                });
            };
        },

        apiImportSchedules(body: ImportSchedulesRequest, options?: any): (fetch?: FetchAPI, basePath?: string) => Promise<ImportSchedulesReply> {
            const localVarFetchArgs = DefaultApiFetchParamCreator(configuration).apiImportSchedules(body, options);
            return (fetch: FetchAPI = window.fetch, basePath: string = BASE_PATH) => {
                return fetch(basePath + localVarFetchArgs.url, localVarFetchArgs.options).then((response) => {
                    if (response.status >= 200 && response.status < 300) {
                        return response.json();
                    } else {
                        throw response;
                    }
                });
            };
        },
    }
};

export const DefaultApiFactory = function (configuration?: Configuration, fetch?: FetchAPI, basePath?: string) {
    return {
        apiGetSerials(options?: any) {
            return DefaultApiFp(configuration).apiGetSerials(options)(fetch, basePath);
        },

        apiImportSchedules(body: ImportSchedulesRequest, options?: any) {
            return DefaultApiFp(configuration).apiImportSchedules(body, options)(fetch, basePath);
        },
    };
};

export class DefaultApi extends BaseAPI {
    public apiGetSerials(options?: any) {
        return DefaultApiFp(this.configuration).apiGetSerials(options)(this.fetch, this.basePath);
    }

    public apiImportSchedules(body: ImportSchedulesRequest, options?: any) {
        return DefaultApiFp(this.configuration).apiImportSchedules(body, options)(this.fetch, this.basePath);
    }
}
