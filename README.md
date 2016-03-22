# Nomi

[![Build Status](https://api.travis-ci.org/giantswarm/nomi.svg)](https://travis-ci.org/giantswarm/nomi)
[![](https://godoc.org/github.com/giantswarm/nomi?status.svg)](http://godoc.org/github.com/giantswarm/nomi)
[![](https://img.shields.io/docker/pulls/giantswarm/nomi.svg)](http://hub.docker.com/giantswarm/nomi)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/nomi)](https://goreportcard.com/report/github.com/giantswarm/nomi)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

**Nomi** is a benchmarking tool that tests a [fleet](https://github.com/coreos/fleet) cluster. With Nomi, you can deploy benchmark units that employ [Docker](https://github.com/docker/docker), [rkt](https://github.com/coreos/rkt) containers, or just raw `systemd` units. Fleemer is able to collect some metrics and generate some plots from those. To make use of Nomi, you just need to define your own benchmark using a YAML file. Nomi parses this file and runs the benchmark according to the instructions defined in it. Additionally, Nomi provides the possibility to define instructions in one line using the parameter `raw-instructions`.

## Requirements

Nomi requires to be installed on a fleet cluster-node to run properly.

Dependencies:

- fleet and systemd running on the host machine.
- In case you want to run Docker or rkt containers, the respective tool needs to be running on the host machines, too.
- Optional: To generate gnu plots, support for gnuplot is required on the host machine. Alternatively, you can run Nomi as a Docker container, which comes with gnuplot installed, as shown below.

## Getting Nomi

Download the latest tarball from here: https://downloads.giantswarm.io/nomi/latest/nomi.tar.gz

Clone the latest git repository from here: https://github.com/giantswarm/nomi.git

Get the latest docker image from here: https://hub.docker.com/r/giantswarm/nomi/

## Running Nomi:

`nomi help`

### Run Nomi from source:

```
make
./nomi help
```

More information on how to run Nomi and its required parameters in: [docs](docs)

## Further Steps

Check more detailed documentation: [docs](docs)

Check code documentation: [godoc](https://godoc.org/github.com/giantswarm/nomi)

## Future Development

- Future directions/vision

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/nomi/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Nomi is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Origin of the Name

`nomi` (のみ[蚤]) is Japanese for [flea](https://en.wikipedia.org/wiki/Flea).
