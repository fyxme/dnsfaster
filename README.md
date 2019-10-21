# dnsfaster

dnsfaster allows you to test the speed and reliability of different DNS servers to check which one is the fastest for your needs.

Faster DNS servers can help improve the speed and reliability of tools used for querying large amounts of DNS records.

dnsfaster was originally developped to find better and faster servers to use while dns bruteforcing.

Additionally, using the fastest dns server can help significantly increase your speed while browsing the internet.

[Why should you test the dns servers you use?](https://gitlab.com/jules.rigaudie/dnsfaster#why-should-you-test-the-dns-servers-you-use)

## dnsfaster in action

Settings:

```
100 workers
100 tests
domain: example.com
output file: out
input file: test-input/10.txt
```

full command: `./dnsfaster -in=test-input/10.txt -workers=100 -tests=100 -domain="example.com" -out=out`

[![asciicast](https://asciinema.org/a/t40ORqxZz5KCB6YXOw8A7H8i9.svg)](https://asciinema.org/a/t40ORqxZz5KCB6YXOw8A7H8i9)

## How does dnsfaster work?

dnsfaster requests multiple random A records from the specified DNS servers.
To generate the random A record, it uses a valid domain and prepends an invalid subdomain.

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

```
$ ./dnsfaster -h
Usage of ./dnsfaster:
  -domain string
    	Domain name to test against (default "example.com")
  -in string
    	The input filepath
  -out string
    	The output file
  -test int
    	Number of tests per dns server (default 100)
  -workers int
    	Number of workers (default 10)
```

Look at `run.sh` for an example command

## Why should you test the dns servers you use?

### Intro

To answer this question we compared efficiency and reliability of massdns ("A high-performance DNS stub resolver for bulk lookups and reconnaissance (subdomain enumeration)") using the best dns servers vs random dns servers.

We used [https://public-dns.info](https://public-dns.info) which has a usefull resources allowing us to get a text file of dns resolvers. ([link](https://public-dns.info/nameservers.txt). This file contained 13219 when we ran this test.

We then ran dnsfaster on that list and find the fastest and most reliable dns resolvers.

We used the following settings:

- number of tests per resolver: 100
- number of workers: 1000

We sorted the output based on average resolve time.

### The resolvers we will be comparing

The top 10 resolvers recorded can be found here (this might differ depending on location):
```
% head fastest.txt -n 10
156.154.70.35
156.154.70.29
64.20.42.252
156.154.70.10
204.246.1.36
156.154.70.14
156.154.70.18
156.154.70.19
216.165.128.161
156.154.70.2
```


The 10 random resolvers we will use for comparaison:
```
% shuf -n 10 fastest.txt
65.78.52.18
124.6.188.176
41.217.204.165
83.218.85.162
213.248.45.60
213.3.18.168
78.111.224.224
92.222.202.244
92.247.8.252
62.2.121.84
```

Before testing the fastest resolvers vs random resolvers, we generate an input file of domains which do not exist in order to not bypass the cache. The script can be found under test notes.

### Massdns outputs

#### fastest resolvers

```
% ./tools/massdns/bin/massdns -r fastest-10.txt -t A -o S -w out in

Progress: 100.00% (00 h 00 min 49 sec / 00 h 00 min 49 sec)
Current incoming rate: 274 pps, average: 20946 pps
Current success rate: 39 pps, average: 20281 pps
Finished total: 1000000, success: 1000000 (100.00%)
Mismatched domains: 32783 (3.17%), IDs: 0 (0.00%)
Failures: 0: 52.80%, 1: 25.72%, 2: 11.68%, 3: 5.24%, 4: 2.39%, 5: 1.14%, 6: 0.54%, 7: 0.24%, 8: 0.12%, 9: 0.06%, 10: 0.03%, 11: 0.02%, 12: 0.01%, 13: 0.00%, 14: 0.00%, 15: 0.00%, 16: 0.00%, 17: 0.00%, 18: 0.00%, 19: 0.00%, 20: 0.00%, 21: 0.00%, 22: 0.00% [...]
```

#### random resolvers

```
% ./tools/massdns/bin/massdns -r randoms-10.txt -t A -o S -w out in

Processed queries: 1000000
Received packets: 1452735
Progress: 100.00% (00 h 02 min 17 sec / 00 h 02 min 17 sec)
Current incoming rate: 2801 pps, average: 10654 pps
Current success rate: 13 pps, average: 7333 pps
Finished total: 1000000, success: 1000000 (100.00%)
Mismatched domains: 452735 (31.16%), IDs: 0 (0.00%)
Failures: 0: 22.85%, 1: 24.89%, 2: 16.71%, 3: 11.93%, 4: 7.91%, 5: 5.29%, 6: 3.52%, 7: 2.36%, 8: 1.59%, 9: 1.05%, 10: 0.67%, 11: 0.45%, 12: 0.27%, 13: 0.18%, 14: 0.11%, 15: 0.08%, 16: 0.05%, 17: 0.03%, 18: 0.02%, 19: 0.01%, 20: 0.01%, 21: 0.00%, 22: 0.00% [...]
```

### Results

Times:

- Random 10 resolvers: 137 seconds
- Fastest 10 resolvers: 49 seconds

137/49 = 2.7959183673 ~= 2.8

- This means that massdns is 2.8 times faster when using the fastest DNS resolvers vs using randomly selected resolvers. 
- The first requests was successfull 52.80% of the time using the fastest and most reliable resolvers while it was only successful 22.85% of the time when using random resolvers.
- The number of mismatched domains is 13.8 times greater when using non random resolvers.

### Conclusion

Using the fastest and most reliable resolvers significantly increased the speed and efficiency of massdns. Therefore showing it is important to do so.

### Additional test notes

- Each test had 1,000,000 DNS queries
- The list of dns queries was using randomly generated sudomain values in order to stop the local DNS server from caching the responses. To generate the subdomains we used the python script below:

```python
import uuid

base_domain = "fyx.me"

number_of_domains = int(1e6) # 1 million

for i in range(number_of_domains):
    print("{}.{}".format(uuid.uuid4().hex,base_domain))
```

and ran it as follows `python generate_domains.py > domains.txt`

- We picked random resolvers from the list of working dns resolvers by running `shuf -n N working-resolvers.txt > resolvers.txt` where N is the number of resolvers requested
- To get the best resolvers we used this command `head -n N working-resolvers.txt > resolvers.txt` where N is the number of resolvers requested

# Limitations

- Does not work for ipv6 yet.
- Will not work if you provide a test domain which responds to all subdomains (wildcard dns entry - \*.domain.com)

# TODO

- add warning when test domain is a wildcard domain
- implement ipv6
- Add test success rate threshold flag (> x%)
- clean up codebase
- silent mode
- add tests
