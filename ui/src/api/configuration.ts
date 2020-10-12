// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

export interface ConfigurationParameters {
    basePath?: string;
}

export class Configuration {
    basePath?: string;

    constructor(param: ConfigurationParameters = {}) {
        this.basePath = param.basePath;
    }
}
