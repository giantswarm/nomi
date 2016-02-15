# We use this image instead of other like Alpine due to bugs with the renderization
FROM ubuntu:14.04

RUN mkdir -p /fleemmer_plots

RUN apt-get update && apt-get install -y gnuplot

# fleemmer
COPY ./fleemmer /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/fleemmer"]
