# Documentation

**Fleemmer** is a benchmarking tool that tests a [fleet](https://github.com/coreos/fleet) cluster. With Fleemmer, you can deploy benchmark units that deploy [Docker](https://github.com/docker/docker), [rkt](https://github.com/coreos/rkt) containers or just raw `systemd` units. Fleemmer is able to collect some metrics and generates some plots. To make use of Fleemmer, you just need to define your own benchmark using a YAML file. Fleemmer parses this file and runs the benchmark according to the instructions defined in it. Additionally, Fleemmer also provides the possibility to define instructions in one line using the parameter `raw-instructions`.

## Fleemmer parameters

- **use-docker**: use benchmark units that deploy [Docker](https://github.com/docker/docker) containers.
- **use-rkt**: use benchmark units that deploy [rkt](https://github.com/coreos/rkt) containers.
- **addr**: address to listen events from the deployed units. This parameter is **important** to allow units notify Fleemmer when they change their state. Fleemmer extracts the public CoreOS IP of the host machine automatically (from `/etc/environment`). Note that you should use this parameter when using a different distro than CoreOS, a Docker container, or a different address to listen on. The `default` port to listen on is `40302`.
- **dump-json**: dump JSON collected metrics to stdout.
- **dump-html-tar**: dump tarred HTML stats to stdout.
- **benchmark-file**: YAML file with the actions to be triggered and the size of the instance groups.
- **raw-instructions**: benchmark raw instructions to be triggered, (requires `instancegroup-size` parameter) and the size of the instance groups.
- **instancegroup-size**: size of the instance group in terms of units, (only if you use `raw-instructions`).
- **generate-gnuplots**: generate gnuplots out of the collected metrics. It is preferable to use `raw-instructions` instead of `benchmark-file` to avoid specifying a docker volume to pass a YAML benchmark definition.
    - **IMPORTANT:** You have to run Fleemmer as a Docker container in your CoreOS machine.

## Benchmark file definition

To start benchmarking our fleet cluster we need to define, which actions our benchmark will perform against a cluster. To do so we can use `raw-instructions` or
`benchmark-file` parameters.

### Using a YAML file with `benchmark-file`

In the following, we detail the purpose of each of the elements that composes a benchmark definition. This file is expected to be a YAML file that follows the format below.

- **instancegroup-size**: indicates the amount of units that will conform an instance group.
- **instructions**: contains a list of instructions that will be executed in descending order. Each instruction can optionally have one of the following elements:
    - **start**:
    	- 	**max**: represents the amount of units to start.
    	- **interval**: in **milliseconds**, it represents the interval of time between start operations.
    - **sleep**: is the amount of time in **seconds** to go to sleep.
    - **float**: **NOT IMPLEMENTED YET** vary the number of units by 'rate' during 'duration' seconds
    	- **rate**: represents the rate of (float).
    	- **duration**: represents the duration in seconds.
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

### Passing a string with the instructions via `raw-instructions`

The main difference is the input format, in which the instructions are entered. When using `raw-instructions`, those are passed in a string fashion manner,
e.g. `--raw-instructions="(sleep 1) (start 200 100) (stop-all)"`. Each parenthesis represents a single instruction that will be executed in sequence and following the inline order. Therefore, a sleep instruction will be followed by a start (with Max: 200 and Duration: 100) and stop operations.

## Running Fleemmer:

Using a benchmark YAML file to run a test that deploys rkt containers:

`fleemmer run --instancegroup-size=1 --dump-json --benchmark-file="./examples/sample01.yaml" --use-rkt`

Using `raw-instructions` and `instancegroup-size` parameters to run a benchmark that deploys raw systemd units:

`fleemmer run --instancegroup-size=1 --dump-json --raw-instructions="(sleep 1) (start 200 100) (sleep 200) (stop-all)"`

### Run Fleemmer from source:

```
make
./fleemmer run --instancegroup-size=1 --dump-json --benchmark-file="./examples/sample01.yaml" --use-rkt
```

Example of a script to send Fleemmer to a remote fleet cluster-node:

```
scp fleemmer core@100.25.10.2:
ssh core@100.25.10.2 'fleemmer run --instancegroup-size=1 --dump-html-tar --benchmark-file="./examples/sample01.yaml"'
```

### Run Fleemmer within a Docker container:

If you want to generate the plots with `gnuplot` in a specific directory `$PLOTS_DIR` use the Docker build:

```
PLOTS_DIR=/tmp
...

docker run -ti \
 -v $PLOTS_DIR:/fleemmer_plots \
 -v /var/run/fleet.sock:/var/run/fleet.sock \
 --net=host \
 --pid=host \
 giantswarm/fleemmer:latest \
 --addr=192.68.10.101:54541 \
 --generate-gnuplots \
 --raw-instructions="(sleep 1) (start 10 100) (sleep 60) (stop-all)"
```


## Collect the results of a benchmark

By default, Fleemmer prints a histogram to stdout that shows the delay of units when starting in the cluster. Additionally, Fleemmer also offers two more options to render the results.

Example of a histogram of starting 900 units in a fleet cluster.

```
1.249-8.583  48.9%   ████████████████████▏  440
8.583-15.92  21.2%   ████████▋              191
15.92-23.25  4.89%   ██▏                    44
23.25-30.59  4.22%   █▊                     38
30.59-37.92  5.22%   ██▏                    47
37.92-45.25  4.78%   ██                     43
45.25-52.59  4.56%   █▉                     41
52.59-59.92  3.33%   █▍                     30
59.92-67.26  2.11%   ▉                      19
67.26-74.59  0.778%  ▍                      7
```

### Dump the colleted metrics

We can either dump the whole metrics as a JSON to stdout, or dump the output into a javascript file that could be used as input to generate d3 graphs. You can find more details in the `output/embedded` directory.

The JSON output follows the next format:

- Start: contains all timestamps and calculated delays of the start operation for each unit.
- Stop: contains all timestamps and calculated delays of the stop operation for each unit.
- EventLog: prints the benchmark instructions that have been launched.
- MachineStates: contains all the data points with the CPU usage for systemd and fleet daemons for each one of the nodes in the fleet cluster.

### Generate gnuplots

For this option, your host should support GNUplot to be able to generate graphs. In CoreOS distros, gnuplot is not installed so we recommend to run Fleemmer as a Docker container there. To do so, you can use this script and pass a directory to collect the plots (at the end).

```
PLOTS_DIR=/tmp

docker run -ti \
 -v $PLOTS_DIR:/fleemmer_plots \
 -v /var/run/fleet.sock:/var/run/fleet.sock \
 --net=host \
 --pid=host \
 giantswarm/fleemmer \
 run
 --addr=192.68.10.101:54541 \
 --generate-plots \
 --raw-instructions="(sleep 1) (start 10 100) (sleep 60) (stop-all)"
```

**NOTE:** We used a heavier base Docker image due to bugs when using the gnuplot package of lighter linux distros like Alpine.

Initially, we just generate four plots but you can also generate your own customized plots with this tool.

Example of a gnuplot that shows the CPU usage(%) of `systemd` at different moments in a fleet cluster:

![systemd](images/systemd.png)

Example of a gnuplot that shows the CPU usage(%) of `fleetd` at different moments in a fleet cluster:

![fleetd](images/fleetd.png)

Example of a gnuplot that shows the delays (seconds) of `start` operations:
![units_start](images/units_start.png)

Example of a gnuplot that shows the delays (seconds) of `stop` operations:
![units_stop](images/units_stop.png)
