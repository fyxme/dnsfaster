#!/bin/sh
#usage: ./dnsfaster <input filepath> <num_workers> <num_tests> <test domain>
rm timed.txt sorted.txt nameservers.txt

wget https://public-dns.info/nameservers.txt

./dnsfaster "nameservers.txt" 2500 100 "ns.agency" "timed.txt" 2> errors.txt

sort -k 2 -n timed.txt -t "," > sorted.txt
tail -n +2 sorted.txt | head -3000 | cut -d, -f1 > fastest.txt
