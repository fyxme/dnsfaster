# dnsfaster

dnsfaster allows you to test different DNS servers to check which one is the fastest for your needs.

## How does dnsfaster work?

It requests multiple random A records from the specified DNS.
To generate the random A record, it uses a valid domain and prepends an invalid subdomain.

# Requirements

- DNS library : `https://github.com/miekg/dns`

To install:
```
go get github.com/miekg/dns
go build github.com/miekg/dns
```

# Usage

`usage: ./dnsfaster <input filepath> <num_workers> <num_tests> <test domain> <out filepath>`

Flags and options coming soon..

# Limitations

- Does not work for ipv6 yet.
- Will not work if you provide a test domain which responds to all subdomains (wildcard dns entry - *.domain.com)
