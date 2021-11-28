# go-clockboundc

[![Test](https://github.com/shogo82148/go-clockboundc/actions/workflows/test.yml/badge.svg)](https://github.com/shogo82148/go-clockboundc/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/shogo82148/go-clockboundc.svg)](https://pkg.go.dev/github.com/shogo82148/go-clockboundc)
[![Coverage Status](https://coveralls.io/repos/github/shogo82148/go-clockboundc/badge.svg?branch=main)](https://coveralls.io/github/shogo82148/go-clockboundc?branch=main)

Golang Client for [ClockBound](https://github.com/aws/clock-bound).

## Run the example

### Prerequisites

[chronyd](https://chrony.tuxfamily.org/) must be running in order to run ClockBoundD.

### Install Rust and Cargo

[Rust](https://www.rust-lang.org/) and [Cargo](https://doc.rust-lang.org/cargo/) are required for building ClockBoundD.
On Linux or another UNIX-like OS, run the following in your terminal.

```bash
curl https://sh.rustup.rs -sSf | sh
```

See [Installation section on The Cargo Book](https://doc.rust-lang.org/cargo/getting-started/installation.html) for more detail.

### Install GCC

You may need to also install gcc:

```
sudo yum install gcc
```

### Install ClockBoundD

Run `cargo install` and start ClockBoundD:

```
cargo install clock-bound-d
$HOME/.cargo/bin/clockboundd
```

If you want to daemonize ClockBoundD, see [Systemd configuration](https://github.com/aws/clock-bound/blob/main/clock-bound-d/README.md#systemd-configuration).

### Run

Now, you can run the example using `go run` command.

```
go run ./cmd/go-clockboundc-now
```

You will get the current system clock with a clock error bound.

```
2021/11/27 10:43:52 Synchronized
2021/11/27 10:43:52 Current:  2021-11-27 10:43:52.95958806 +0000 UTC
2021/11/27 10:43:52 Earliest: 2021-11-27 10:43:52.959488228 +0000 UTC
2021/11/27 10:43:52 Latest:   2021-11-27 10:43:52.959687892 +0000 UTC
2021/11/27 10:43:52 Range:    199.664µs
```

## See Also

- [ClockBound](https://github.com/aws/clock-bound)
- [ClockBound Protocol Version 1](https://github.com/aws/clock-bound/blob/main/PROTOCOL.md)
- [Amazon Time Sync Service now makes it easier to generate and compare timestamps](https://aws.amazon.com/about-aws/whats-new/2021/11/amazon-time-sync-service-generate-compare-timestamps/)
