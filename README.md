# dnsfaster

dnsfaster allows you to test the speed and reliability of different DNS servers to check which one is the fastest for your needs.

Faster DNS servers can help improve the speed and reliability of tools used for querying large amounts of DNS records.

dnsfaster was originally developped to find better and faster servers to use while dns bruteforcing.

Additionally, using the fastest dns server can help significantly increase your speed while browsing the internet.


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

`usage: ./dnsfaster <input filepath> <num_workers> <num_tests> <test domain> <out filepath>`

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

