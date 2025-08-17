# DNS Bench

A fast and flexible DNS resolver benchmarking tool written in Go. Compare the performance of multiple DNS resolvers with detailed statistics and optional CSV export.

## Features

- **Multiple Resolver Support**: Test multiple DNS resolvers simultaneously
- **Detailed Statistics**: Min, Max, Average, Median, and 95th percentile response times
- **Cold/Warm Testing**: Option to test with cache-busting random subdomains
- **IPv4/IPv6 Support**: Test both A and AAAA record lookups
- **CSV Export**: Export detailed results for further analysis
- **Configurable Timeouts**: Set custom timeout values for queries
- **Error Reporting**: Track and report DNS resolution errors

## Installation

### Pre-built Binary
Download the latest release from the [releases page](https://github.com/ohidurbappy/dns-bench/releases).

### Build from Source
```bash
git clone https://github.com/ohidurbappy/dns-bench.git
cd dns-bench
go build -o dnsbench main.go
```

## Usage

### Basic Usage
```bash
./dnsbench -domain example.com
```

### Advanced Usage
```bash
./dnsbench \
  -domain google.com \
  -count 20 \
  -timeout 2s \
  -network ip4 \
  -cold \
  -resolvers "Cloudflare=1.1.1.1,Google=8.8.8.8,Quad9=9.9.9.9" \
  -out results.csv
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-domain` | `example.com` | Domain name to resolve |
| `-count` | `10` | Number of queries per resolver |
| `-timeout` | `1500ms` | Per-query timeout (e.g., 1500ms, 2s) |
| `-network` | `ip4` | Network type: `ip4` (A records) or `ip6` (AAAA records) |
| `-cold` | `false` | Use random subdomains to bypass resolver cache |
| `-resolvers` | See below | Comma-separated list of Name=IP pairs |
| `-out` | | Optional path to write CSV results |

### Default Resolvers
```
Cloudflare=1.1.1.1
Google=8.8.8.8
Quad9=9.9.9.9
OpenDNS=208.67.222.222
AdGuard=94.140.14.14
```

## Examples

### Test Popular DNS Resolvers
```bash
./dnsbench -domain github.com -count 15
```

### Cold Cache Testing
Test with cache-busting to measure resolver performance without cached results:
```bash
./dnsbench -domain example.com -cold -count 20
```

### IPv6 Testing
```bash
./dnsbench -domain google.com -network ip6 -count 10
```

### Custom Resolvers
```bash
./dnsbench \
  -domain cloudflare.com \
  -resolvers "CF-Primary=1.1.1.1,CF-Secondary=1.0.0.1,Google-Primary=8.8.8.8" \
  -count 25
```

### Export Results to CSV
```bash
./dnsbench -domain example.com -out benchmark_results.csv
```

## Sample Output

```
DNS Benchmark
Target: example.com | Runs: 10 | Timeout: 1.5s | Network: ip4 | Mode: WARM
--------------------------------------------------------------------------------
Resolver      Min    Avg    Med    p95    Max    Success%
--------------------------------------------------------------------------------
Cloudflare    12.3ms 15.7ms 14.2ms 22.1ms 28.4ms    100.0%
Google        18.9ms 23.4ms 21.8ms 31.2ms 35.7ms    100.0%
Quad9         25.1ms 29.8ms 28.3ms 38.9ms 42.1ms    100.0%
OpenDNS       31.2ms 36.7ms 35.1ms 45.8ms 48.9ms    100.0%
AdGuard       28.7ms 33.2ms 31.9ms 41.3ms 44.6ms    100.0%
```

## CSV Output Format

The CSV export includes two sections:

### Summary Statistics
- Resolver name
- Query count and success count
- Response time statistics (min, avg, median, p95, max) in milliseconds
- Error messages (if any)

### Individual Query Results
- Resolver name
- Run index
- Individual query duration in milliseconds
- Error message (if query failed)

## Use Cases

- **Network Performance Testing**: Compare DNS resolver performance from your location
- **Infrastructure Planning**: Choose the best DNS resolver for your applications
- **Troubleshooting**: Identify DNS resolution issues and performance bottlenecks
- **Monitoring**: Regular benchmarking to track DNS performance over time
- **Research**: Analyze DNS resolver behavior under different conditions

## Technical Details

- Uses Go's `net.Resolver` with UDP transport
- Supports custom resolver ports (format: `Name=IP:Port`)
- Implements proper timeout handling and error reporting
- Calculates statistical measures including percentiles
- Thread-safe concurrent execution

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go's standard library
- Inspired by the need for simple, effective DNS benchmarking tools