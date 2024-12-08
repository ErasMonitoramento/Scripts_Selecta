#!/bin/bash

ip=$1
porta=$2
user=$3
pass=$4
ifname=$5
cmd="display pppoe statistics interface GigabitEthernet 0/7/0"

sshpass -p $pass ssh -o LogLevel=quiet -p $porta $user@$ip $cmd 2> /dev/null | grep 'TOTAL' | awk '{print $3}'