#!/usr/bin/python3
# coding: utf-8

import subprocess


def say_hello():
	print("\n Mapago evaluation script here")

def call_mapago():
	args = ("/home/fronhoef/GoProgramming/bin/mapago-client", "-ctrl-addr", "169.254.206.129", "-ctrl-protocol", "tcp", "-msmt-type", "tcp-throughput", "-msmt-streams", "10", "-msmt-listen-addr", "169.254.206.129")
	popen = subprocess.Popen(args, stdout=subprocess.PIPE)
	popen.wait()
	output = popen.stdout.read()
	print(output)


if __name__ == '__main__':
	say_hello()
	call_mapago()