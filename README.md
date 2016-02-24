# Fleemmer

[![Build Status](https://api.travis-ci.org/giantswarm/fleemmer.svg)](https://travis-ci.org/giantswarm/fleemmer)
[![](https://godoc.org/github.com/giantswarm/fleemmer?status.svg)](http://godoc.org/github.com/giantswarm/fleemmer)
[![](https://img.shields.io/docker/pulls/giantswarm/fleemmer.svg)](http://hub.docker.com/giantswarm/fleemmer)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

**Fleemmer** is a benchmarking tool that tests a [fleet](https://github.com/coreos/fleet) cluster. With Fleemmer, you can deploy benchmark units that deploy [Docker](https://github.com/docker/docker), [rkt](https://github.com/coreos/rkt) containers or just raw `systemd` units. Fleemer is able to collect some metrics and generates some plots. To make use of Fleemmer, you just need to define your own benchmark using a YAML file. Fleemmer parses this file and runs the benchmark according to the instructions defined in it. Additionally, Fleemmer also provides the possibility to define instructions in one line using the parameter `raw-instructions`.

## Requirements

Fleemmer requires to be installed on a fleet cluster-node to run properly.

Dependencies:

- fleet, docker, rkt and systemd running on the host machine.
- Optional: To generate gnu plots, support for gnuplot is required on the host machine. Alternatively, you can run Fleemmer as a Docker container, which comes with gnuplot installed, as shown below.

## Getting Fleemmer

Download the latest tarball from here: https://downloads.giantswarm.io/fleemmer/latest/fleemmer.tar.gz

Clone the latest git repository version from here: `git@github.com:giantswarm/fleemmer.git`

Get the latest docker image from here: https://hub.docker.com/r/giantswarm/fleemmer/

## Running Fleemmer:

`fleemmer help`

### Run Fleemmer from source:

```
make
./fleemmer help
```

More information on how to run Fleemmer and its required parameters in: [docs](docs)

## Further Steps

Check more detailed documentation: [docs](docs)

Check code documentation: [godoc](https://godoc.org/github.com/giantswarm/fleemmer)

## Future Development

- Future directions/vision

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/fleemmer/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Fleemmer is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
