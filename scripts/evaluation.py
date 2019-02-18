#!/usr/bin/python3
# coding: utf-8

import matplotlib.pyplot as plt
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
print("json lines", lines_json)

# we got one lines of json, convert everyline
# into our "db"
db = []
for line_json in lines_json.splitlines():
    db.append(json.loads(line_json))
# debug check: print now python data
# pp.pprint(db)

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
prev_datetime_max = datetime.datetime(1, 1, 1)
prev_datetime_max_set = False

# we store the data for plot, where just one overall
# bandwith is not succiently,
normalized = []


# now find the youngest and oldest date
for entry in db:

    bytes_measurement_point = 0

    # one entry can have multiple streams, so iterate over the
    # streams now
    for stream in entry:
        time = datetime.datetime.strptime(
            stream['ts-start'], '%Y-%m-%dT%H:%M:%S.%f')
        if time < datetime_min:
            datetime_min = time
            
            if prev_datetime_max_set == False:
            	prev_datetime_max = datetime_min
            	prev_datetime_max_set = True 

        time = datetime.datetime.strptime(
            stream['ts-end'], '%Y-%m-%dT%H:%M:%S.%f')
        if time > datetime_max:
            datetime_max = time

        bytes_measurement_point += int(stream['bytes'])

    # we got data from all streams of that entry
    curr_msmt_time = (datetime_max - datetime_min).total_seconds()
    bytes_per_period = bytes_measurement_point - bytes_rx
    mbits_per_period = (bytes_per_period * 8) / 10**6
    # bytes_rx == # bytes until now received
    bytes_rx = bytes_measurement_point

    # this works only if data is send immediately after prev_datetime_max
    # or we pay attention to a period where nothing is transmitted
    duration_of_period = (datetime_max - prev_datetime_max).total_seconds()
    prev_datetime_max = datetime_max
    throughput_of_period = mbits_per_period / duration_of_period
    normalized.append([curr_msmt_time, throughput_of_period])


measurement_length = (datetime_max - datetime_min).total_seconds()
bytes_sec = bytes_rx / measurement_length
Mbits_sec = (bytes_sec * 8) / 10**6
Kbits_sec = (bytes_sec * 8) / 10**3
print('overall bandwith: {} bytes/sec'.format(bytes_sec))
print('overall bandwith: {} Mbits/sec'.format(Mbits_sec))
print('overall bandwith: {} Kbits/sec'.format(Kbits_sec))
print('measurement length: {} sec]'.format(measurement_length))
print('received: {} bytes]'.format(bytes_rx))

# now plotting starts, not really fancy

# normalize date to start with just 0 sec and not 2018-01-23 ...
x = []
y = []
for i in normalized:
    x.append(i[0])
    y.append(i[1])

plt.plot(x, y)
plt.ylabel('Throughput [MBits/s]')
plt.xlabel('Time [seconds]')
plt.show()
