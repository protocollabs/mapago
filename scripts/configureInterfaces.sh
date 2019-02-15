#!/bin/bash

echo "Configuring interface: $1"

ethtool -K $1 tso off
ethtool -K $1 ufo off
ethtool -K $1 gso off
ethtool -K $1 lro off
ethtool -K $1 gro off