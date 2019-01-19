#!/usr/bin/python3
# coding: utf-8

import sys
import json
import datetime
import pprint
# for debugging, see pp later
pp = pprint.PrettyPrinter(indent=4)

# the script expect the following input format:
# for one stream and one measurement point it looks like:
#   echo -n '[ { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:43.763987", "bytes" : "100"}  ]' | ./scripts/evaluation.py
# For two streams the data looks:
#
# [ { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:43.763987", "bytes" : "100"}, { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:43.763987", "bytes" : "100"}  ]
# [ { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:44.763987", "bytes" : "200"}, { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:43.763987", "bytes" : "200"}  ]
# [ { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:45.763987", "bytes" : "300"}, { "ts-start" : "2017-12-16T12:32:42.763987", "ts-end" : "2017-12-16T12:32:43.763987", "bytes" : "300"}  ]
#
# FIXME: please remove the examples abote, because not proper aligned.
# if it works, all toy comments here can be removed.

# read from pipe (stdin) until end of program
lines_json = sys.stdin.read()

# we got one lines of json, convert everyline
# into our "db"
db = []
for line_json in lines_json.splitlines():
    db.append(json.loads(line_json))
# debug check: print now python data
#pp.pprint(db)

if len(db) < 1:
    print("nope, need at least ome measurement to calculate delta")
    sys.exit(1)

# min and max will at the end of the next loop contain
# the real min and max values
datetime_min = datetime.datetime(4000, 1, 1)
datetime_max = datetime.datetime(1, 1, 1)

# Keep in mind for all byte accounting here:
# it is transmitted in a ABSOLUTE manner, no differences
# are signaled via mapago
bytes_rx = 0

# we store the data for plot, where just one overall
# bandwith is not succiently, 
normalized = []


# now find the youngest and oldest date
for entry in db:

    bytes_measurement_point = 0

    # one entry can have multiple streams, so iterate over the
    # streams now
    for stream in entry:
        time = datetime.datetime.strptime(stream['ts-start'], '%Y-%m-%dT%H:%M:%S.%f')
        if time < datetime_min:
            datetime_min = time
        time = datetime.datetime.strptime(stream['ts-end'], '%Y-%m-%dT%H:%M:%S.%f')
        if time > datetime_max:
            datetime_max = time

        bytes_measurement_point += int(stream['bytes'])


    # now account data for plotting
    normalized.append([datetime_max, bytes_measurement_point - bytes_rx])
    bytes_rx = bytes_measurement_point

measurement_length = (datetime_max - datetime_min).total_seconds()
bytes_sec = bytes_rx / measurement_length
print('overall bandwith: {} bytes/sec'.format(bytes_sec))
print('measurement length: {} sec]'.format(measurement_length))
print('received: {} bytes]'.format(bytes_rx))


# now plotting starts, not really fancy
import matplotlib.pyplot as plt

# normalize date to start with just 0 sec and not 2018-01-23 ...
x = []; y = []
for i in normalized:
    x.append((i[0] - datetime_min).total_seconds())
    y.append(i[1])

plt.plot(x, y)
plt.ylabel('Throughput [bytes/s]')
plt.xlabel('Time [seconds]')
plt.show()
