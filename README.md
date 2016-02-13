# Fleemmer

[![Build Status](https://api.travis-ci.org/giantswarm/fleemmer.svg)](https://travis-ci.org/giantswarm/fleemmer)
[![](https://godoc.org/github.com/giantswarm/fleemmer?status.svg)](http://godoc.org/github.com/giantswarm/fleemmer)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

**Fleemmer** is a benchmarking tool that tests a [fleet](https://github.com/coreos/fleet) cluster, fleemer is able to collect some metrics and generates some plots. To make use of Fleemmer, your just need to define your own benchmark using a YAML file. Fleemmer parses this file and runs the benchmark according to the instructions defined in it. Additionally, Fleemmer also provides the possibility to define instructions in one line using the parameter `raw-instructions`.

## Requirements

Fleemmer requires to be installed on a fleet cluster-node to run properly.

Dependencies:

- fleet and systemd running on the host machine.
- Optional: To generate gnu plots, it is required to have support for gnuplot in the host machine. Otherwise you could run Fleemmer as a docker container to have gnuplot installed, as shown below.

## Benchmark file definition

In the following, we detail the purpose of each one of the elements that composes a benchmark definition.

- **instancegroup-size**: indicates the amount of units that will conform an instance group.
- **instructions**: contains a list of instructions that will be executed in descending order. Each instruction can optionally have one of the following elements:
    - **start**:
    	- 	**max**: represents the amount of units to start.
    	- **interval**: In **Milliseconds**, it represents the interval of time between start operations.
    - **sleep**: is the amount of time in **Seconds** to go to sleep.
    - **float**:
    	- **rate**: represents the rate of (float).
    	- **duration**: represents the duration in Seconds.
    - **expect-running**:
	    - **amount**: represents the amount of expected running units.
    	- **symbol**: used to indicate whether you expect `[<|>]` `expect-running/amount` units to be running.
    - **stop**: indicates the directive used to stop the current units (stop-all|). At this moment, we only offer `stop-all` as an alternative to stop units.

    **NOTE:** The order of the elements in an instruction indicates, in which order such an action will be triggered.

**Example:**

```
instancegroup-size: 1
instructions:
  - start:
     max: 8
     interval: 200
  - expect-running:
    symbol: <
    amount: 10
  - sleep: 10
  - start:
     max: 3
     interval: 300
  - sleep: 200
  - stop: stop-all
```

## Fleemmer parameters

- **addr**: address to listen.
- **dump-json**: dump JSON stats to stdout.
- **dump-html-tar**: dump tarred HTML stats to stdout.
- **benchmark-file**: YAML file with the actions to be triggered and the size of the instance groups.
- **raw-instructions**: benchmark raw instructions to be triggered, (requires `instancegroup-size` parameter) and the size of the instance groups.
- **instancegroup-size**: size of the instance group in terms of units, (only if you use `raw-instructions`).
- **generate-gnuplots**: generate gnuplots out of the collected metrics. It is preferable to use `raw-instructions` instead of `benchmark-file` to avoid specifying a docker volume to pass a YAML benchmark definition.
    - **IMPORTANT:** You have to run Fleemmer as a Docker container in your CoreOS machine.

## Run Fleemmer:

Using a benchmark YAML file to run a test:

`./fleemmer -addr=100.25.10.2:54541 -v=12 -instancegroup-size=1 -dump-html-tar -benchmark-file="./examples/sample01.yaml" &>> outputFile `

Using `raw-instructions` and `instancegroup-size` parameters to run a benchmark:

`./fleemmer -addr=192.68.10.102:54541 -v=12 -instancegroup-size=1 -dump-json -raw-instructions="(sleep 1) (start 200 100) (sleep 200) (stop-all)" &>>outfile`

Example of a script to send Fleemmer to a remote fleet cluster-node:

```
scp fleemmer core@100.25.10.2:
ssh core@100.25.10.2 './fleemmer -addr=100.25.10.2:54541 -v=12 -instancegroup-size=1 -dump-html-tar -benchmark-file="./examples/sample01.yaml"'
```

If you want to generate the plots with `gnuplot` in a specific directory `$PLOTS_DIR` use the Docker build:

```
PLOTS_DIR=/tmp
...

docker run -ti \
 -v $PLOTS_DIR:/fleemmer_plots \
 -v /var/run/fleet.sock:/var/run/fleet.sock \
 --net=host \
 --pid=host \
 giantswarm/fleemmer \
 -addr=192.68.10.101:54541 \
 -generate-gnuplots \
 -raw-instructions="(sleep 1) (start 10 100) (sleep 60) (stop-all)"
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/fleemmer/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Fleemmer is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
