# Documentation

**Fleemmer** is a benchmarking tool that tests a [fleet](https://github.com/coreos/fleet) cluster, fleemer is able to collect some metrics and generates some plots. To make use of Fleemmer, your just need to define your own benchmark using a YAML file. Fleemmer parses this file and runs the benchmark according to the instructions defined in it. Additionally, Fleemmer also provides the possibility to define instructions in one line using the parameter `raw-instructions`.

## Benchmark file definition

To start benchmarking our fleet cluster we need to define which actions our benchmark will perform against a cluster. To do so we can use `raw-instructions` or
`benchmark-file` parameters.

### Using a YAML file with `benchmark-file`

In the following, we detail the purpose of each one of the elements that composes a benchmark definition. This file is expected to be a YAML file that follows
the next format.

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
  - sleep: 1
  - start:
     max: 2700
     interval: 100
  - sleep: 700
  - stop: stop-all
```

### Passing a string with the instructions via `raw-instructions`

The main difference is that input format on which the instructions are entered. When using `raw-instructions`, those are passed in a string fashion manner,
e.g. `-raw-instructions="(sleep 1) (start 200 100) (stop-all)"`. Each parenthesis represents a single instruction that will be executed in sequence and following the inline order. Therefore, a sleep instruction will be followed by a start (with Max: 200 and Duration: 100) and stop operations.


## Collect the results of my benchmark

By default, Fleemmer prints in the stdout an histogram that shows the delay of units when starting in the cluster. Additionally to this, Fleemmer also offers two more options to render the results.

Example of an histogram, when starting 900 units in a fleet cluster.

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

We can either dump the whole metrics as a JSON to stdout, or to dump the output into a javascript file that could be used as input to generate d3 graphs. You can find more details in `output/embedded` directory.

The JSON output follows the next format:

- Start: contains all the timestamp and calculated delay of the start operation for each unit.
- Start: contains all the timestamp and calculated delay of the start operation for each unit.
- EventLog: prints the benchmark instructions that have been launched.
- MachineStates: contains all the data points with the CPU usage for systemd and fleet daemons for each one of the nodes in the fleet cluster.

### Generate gnuplots

For this option, we recommend to support GNUplot software to be able to generate graphs. In CoreOS distros, gnuplot is not installed so we then recommend to run Fleemmer as a Docker container. To do so, you can use this script and pass a directory to collect the plots (at the end).

```
PLOTS_DIR=/tmp

docker run -ti \
 -v $PLOTS_DIR:/fleemmer_plots \
 -v /var/run/fleet.sock:/var/run/fleet.sock \
 --net=host \
 --pid=host \
 giantswarm/fleemmer \
 -addr=192.68.10.101:54541 \
 -generate-plots \
 -raw-instructions="(sleep 1) (start 10 100) (sleep 60) (stop-all)"

```

**NOTE:** We used a heavier base docker image due to bugs when using the gnuplot package of lighter linux distros like Alpine.

Initially, we just generate four plots but we believe more a better plots can be generated with this tool.

Example of a gnuplot, it shows the CPU usage(%) of `systemd` at different moments in a fleet cluster:

![systemd](https://cloud.githubusercontent.com/assets/3602792/13027517/b877cee0-d252-11e5-97a9-fcf18ffc802d.png)

Example of a gnuplot, it shows the CPU usage(%) of `fleetd` at different moments in a fleet cluster:

![fleetd](https://cloud.githubusercontent.com/assets/3602792/13027516/b871b064-d252-11e5-8d02-461f3b0b49f9.png)


Example of a gnuplot, it shows the delays(seconds) of `start` operations:
![units_start](https://cloud.githubusercontent.com/assets/3602792/13027491/48f5704a-d252-11e5-83bf-1fec97bc953f.png)

Example of a gnuplot, it shows the delays(seconds) of `stop` operations:
![units_stop](https://cloud.githubusercontent.com/assets/3602792/13027518/b87e3f28-d252-11e5-8768-416b7c379c5a.png)
