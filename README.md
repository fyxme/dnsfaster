# dnsfaster

dnsfaster allows you to test the speed and reliability of different DNS servers to check which one is the fastest for your needs.

Faster DNS servers can help improve the speed and reliability of tools used for querying large amounts of DNS records.

dnsfaster was originally developped to find better and faster servers to use while dns bruteforcing.

Additionally, using the fastest dns server can help significantly increase your speed while browsing the internet.

## dnsfaster in action

Settings:
```
100 workers
100 tests
domain: example.com
output file: out
input file: test-input/10.txt
```

full command: `./dnsfaster test-input/10.txt 100 100 "example.com" out` 

[![asciicast](https://asciinema.org/a/t40ORqxZz5KCB6YXOw8A7H8i9.svg)](https://asciinema.org/a/t40ORqxZz5KCB6YXOw8A7H8i9)

## How does dnsfaster work?

dnsfaster requests multiple random A records from the specified DNS servers.
To generate the random A record, it uses a valid domain and prepends an invalid subdomain.

## Why should you test the dns servers you use?

To answer this question we compared efficiency and reliability of masscan ("A high-performance DNS stub resolver for bulk lookups and reconnaissance (subdomain enumeration)") using the best dns servers vs random dns servers.

We used [https://public-dns.info](https://public-dns.info) which has a usefull resources allowing us to get a text file of dns resolvers. ([link](https://public-dns.info/nameservers.txt). This file contained 13219 when we ran this test.

We then ran dnsfaster on that list and find the fastest and most reliable dns resolvers.

We used the following settings:
- number of tests per resolver: 100
- number of workers: 1000

We sorted the output based on average resolve time.

The top 5 resolvers recorded can be found here (this might differ depending on location):
```
156.154.70.35
156.154.70.29
64.20.42.252
156.154.70.10
204.246.1.36
```

We tested massdns with the best resolvers vs using random resolvers and varied the number of used resolvers. The results can be seen in the table below.

.....
.....
.....
//TODO HERE
.....
.....
.....


Test notes:

- Each test had 1,000,000 DNS queries 
- The list of dns queries was using randomly generated sudomain values in order to stop the local DNS server from caching the responses. To generate the subdomains we used the python script below:
```python3
import uuid

base_domain = "fyx.me"

number_of_domains = int(1e6) # 1 million

for i in range(number_of_domains):
    print("{}.{}".format(uuid.uuid4().hex,base_domain)
```
and ran it as follows `python generate_domains.py > domains.txt`

- We picked random resolvers from the list of working dns resolvers by running `shuf -n N working-resolvers.txt > resolvers.txt` where N is the number of resolvers requested
- To get the best resolvers we used this command `head -n N working-resolvers.txt > resolvers.txt` where N is the number of resolvers requested


# Requirements

- DNS library : `https://github.com/miekg/dns`

To install required library run:
```
go get
```

To build the binary:
```
go build dnsfaster.go
```

# Usage

`usage: ./dnsfaster <input filepath> <num_workers> <num_tests> <test domain> <out filepath>`

Look at `run.sh` for an example command

Flags and options coming soon..

# Limitations

- Does not work for ipv6 yet.
- Will not work if you provide a test domain which responds to all subdomains (wildcard dns entry - \*.domain.com)

# TODO

- add flags and options
- add warning when test domain is a wildcard domain
- implement ipv6
- Add test success rate threshold flag (> x%)
- clean up codebase

