# Maparo

## Current installation procedure

You can use pre-compiled releases compiled for Linux, Mac and Windows.

> Note: the feature set is not identical on all platforms. The module
> help describe all pitfalls.

Execute the following installation steps to use Mapago:

```
export GOPATH=$HOME/go
export GOBIN=$HOME/go/bin
go get github.com/protocollabs/mapago
cd $GOPATH/github.com/protocollabs/mapago
make install
mapago-server
mapago-client -ctrl-addr 169.254.206.129 -ctrl-protocol tcp -msmt-type tcp-throughput
```

## Design Principles

- Modules can be used in parallel with other modules. Modules can be initiated
  several times. This allows more complex campaigns like *Realtime Response Under
  Load (RRUL)* like FLENT is doing.
- *Maparo* return either a JSON set or human formatted data. *Maparo* do
  not draw or visualize any data. To visualize data you can use Python
  scripts using the powerful and flexible matplotlib library.
- *Maparo* do not focus on collection OS internal stats for further analysis
  for now. This can later be added.

## Using Maparo from Third Party Tools/Scripts Automatically

Maparo will exit with return code 0 if everything went fine. If not return code 1 signals
the script that something went wrong. To communicate error, warning and debugging
information to the outside *maparo* will use STDERR. STDOUT is just use for analysis
data. This a script can trust that the outputed string is a pure JSON string (if enabled)
and not garbaged with other information.

```
maparo <args> | python3 graph-script.py -o output.png
```

## Parsing Format

*Maparo* will print out the results and the input parameters as well as other
gathered system parameters if JSON is the output format:

```
{
  "version" = "semver value",
  "system" = {
  },
  "input" = {
    // flaten config, if only one module is
    // used the module config including name is
    // printed. If a campaign is used the serialized
    // campaign included config values is printed 
    "mod-tcp-goodput" = {
    },
  },
  "output" : {
  },
}
```

## Development

### Install binary via

```
go get github.com/protocollabs/maparo/...
```

### Prepare Development Environment

For Debian based systems install go suite:

```
sudo apt-get install golang-go
```


# Links and Referrences

## Go Networking Examples

- https://github.com/golang/net/blob/master/ipv4/example_test.go

## Plotting Ideas

If the JSON output includes information what module/campaign
was called and all information is available (i.e. remote call,
client and server json in one json structure it is possible to
pipe the data to a python3 script. The Python3 script will
visualize and output the information.

```
# server
maparo daemon

# client
mapago module tcp-througput format=json conf.dst=::1 | mapago-chart --output-dir tcp-througput

$ ls tcp-througput
tcp-througput-over-time.data
tcp-througput-over-time.py
tcp-througput-over-time.png

tcp-througput-info.data
tcp-througput-info.py
tcp-througput-info.png

[...]
```

- http://blog.enthought.com/general/visualizing-uncertainty/
- https://plot.ly/python/box-plots/
- http://jonchar.net/notebooks/matplotlib-styling/
- http://people.duke.edu/~ccc14/pcfb/numpympl/MatplotlibBarPlots.html
