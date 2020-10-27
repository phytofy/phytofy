# OSRAM PHYTOFY® RL

PHYTOFY® RL is a tunable LED horticultural lighting system from UV to far-red with real-time control and scheduling features for each individual channel. The calibrated system is designed for delivering light treatments with varying spectra, wavelength and intensity, for horticulture research.

More information can be found at [www.osram.com/phytofy](https://www.osram.com/phytofy).


## Open Protocol Specification

* [Documentation - Protocol Specification & System Integration](docs/ProtocolSpecificationAndSystemIntegration.md)


## Reference Implementation

### Introduction

This repository contains an example code which implements the aspects covered by the [Protocol Specification & System Integration](docs/ProtocolSpecificationAndSystemIntegration.md) document with a 1-to-1 mapping between the functionality which can be accessed over that protocol and functionality exposed by this code. As such, it serves only as illustration rather than a production-grade software.

In short, the functions of the protocol allow to list fixtures, set the LED channels' levels, as well as create/delete schedules and pause/resume scheduling.

The added value of that example layer of code is best described by the following list:

* The code monitors periodically the network for Moxa NPort® adapters on the network and fixtures behind them, and abstracts away working with adapters and short RS485 addresses. Instead, it offers an interface with addressing of fixtures directly by their serial numbers.
* The functions are exposed as a CLI (Command Line Interface) application – packaged as Windows executable (amd64 architecture)
* The functions are exposed as an [OpenAPI](api/hw1.yaml) application – packaged as a Docker container (amd64 & arm32v7 architectures); Note: Please mind that the API does not follow the best practices of CRUD mapping.


### Command Line Interface

The CLI can be build and deposited to current working directory with the following commands:

    ./automation/build.cli.sh
    ./automation/publish.cli.sh

Here is an example of how to set LED channels' levels with the CLI (here: green channel at 100%):

    phytofy.exe v1-set-leds-pwm '{"serial": 206001, "payload": {"config": 3, "levels": [0, 0, 100, 0, 0, 0]}}'

Running the application without any arguments will list all available commands. To learn the arguments expected by each command please see the [OpenAPI](api/hw1.yaml).


### OpenAPI & UI

With the following command one can build the Docker container images for arm64 and arm32v7 architectures:

    ./automation/build.iot.sh

The resulting Docker images will be tagged `phytofy-amd64:latest` and `phytofy-arm32v7:latest`.

To run the OpenAPI locally the following commands can be used:

    docker run -d --network host -p 8080:8080 phytofy-amd64:latest v1-api 8080
    docker run -d --network host -p 8080:8080 phytofy-arm32v7:latest v1-api 8080

Note that the `--network host` is necessary for the application to be able to communicate with fixtures.

Running one of these commands will expose the API on the local device on the port 8080. As an example, the CLI command `v1-set-leds-pwm` corresponds to the [OpenAPI](api/hw1.yaml) path `/v1/set-leds-pwm` and the JSON-formatted argument to the body of the request defined by that specification.

To use the OpenAPI to set LED channels' levels run a command like this:

    curl -X POST -H "Content-Type: application/json" --data '{"serial": 206001, "payload": {"config": 3, "levels": [0, 0, 100, 0, 0, 0]}}' http://localhost:8080/v1/set-leds-pwm

A small subset of the functionality ([schedule](docs/ui/01.Lighting_Schedules_List.png) [editing](docs/ui/02.Lighting_Schedule_Editing.png) & [irradiance map simulation](docs/ui/02.Lighting_Schedule_Editing.png)) is also exposed as a UI. To access the UI run one of the following commands and go to `http://localhost:8080/`:

    docker run -d --network host -p 8080:8080 phytofy-amd64:latest v1-app 8080
    docker run -d --network host -p 8080:8080 phytofy-arm32v7:latest v1-app 8080


### Scheduling

Both the CLI (command `v1-import-schedules`) as well as the UI allow to import and apply schedules from a CSV file. Here is an example contents of such file:

```
2020-09-12,2019-10-15,09:00,17:00,0,50,50,50,50,50,100300
2020-09-12,2019-10-15,09:00,17:00,0,10,10,10,10,10,100400
```

The semantics is as follows:

```
StartDate,StopDate,StartTime,StopTime,UVA,Blue,Green,HyperRed,FarRed,White,SerialNumber
```


### Logging

By setting the PHYTOFY_CONSOLE_LOGGING environemnt variable to `true` the application will output logs directly to console. Otherwise the logs will be stored in `logs` subdirectory of the directory where the application resides.

Log rotation scheme is in place with a 64 day long retention period and 100 MB maximum log file limit.

If the application runs as a docker container the logs can be stored on the host machine by adding a volume mapping - e.g. `-v $PWD:/logs`:

    docker run -d --network host -p 8080:8080 -v $PWD:/logs phytofy-amd64:latest v1-app 8080
    docker run -d --network host -p 8080:8080 -v $PWD:/logs phytofy-arm32v7:latest v1-app 8080


## OSRAM Horticultural Lighting

* [www.osram.com/horticulture](https://www.osram.com/horticulture)


## Licenses

This repository uses the [MIT](LICENSE) license for code.
