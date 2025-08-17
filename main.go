package main

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

type Sample struct {
	Duration time.Duration
	Err      error
}

type Stats struct {
	Count       int
	Successes   int
	Min         time.Duration
	Max         time.Duration
	Avg         time.Duration
	Median      time.Duration
	P95         time.Duration
	Errors      []error
	DurationsMs []float64
}

type ResolverCfg struct {
	Name string
	Addr string // host or host:port (port defaults to 53 if omitted)
}

type Row struct {
	Name    string
	Stats   Stats
	Samples []Sample
}

func main() {
	domain := flag.String("domain", "example.com", "Domain to resolve")
	count := flag.Int("count", 10, "Number of queries per resolver")
	timeout := flag.Duration("timeout", 1500*time.Millisecond, "Per-query timeout (e.g. 1500ms, 2s)")
	network := flag.String("network", "ip4", "Network: ip4 or ip6 (A vs AAAA)")
	cold := flag.Bool("cold", false, "Cold mode: use random subdomain each query to bust resolver cache")
	resolversCSV := flag.String("resolvers", "Cloudflare=1.1.1.1,Google=8.8.8.8,Quad9=9.9.9.9,OpenDNS=208.67.222.222,AdGuard=94.140.14.14", "Resolvers as Name=IP[,Name=IP...]")
	outCSV := flag.String("out", "", "Optional path to write CSV results")
	flag.Parse()

	resolvers := parseResolvers(*resolversCSV)
	if len(resolvers) == 0 {
		fmt.Println("No resolvers provided.")
		os.Exit(1)
	}

	fmt.Printf("DNS Benchmark\n")
	fmt.Printf("Target: %s | Runs: %d | Timeout: %v | Network: %s | Mode: %s\n",
		*domain, *count, *timeout, *network, ternary(*cold, "COLD", "WARM"))
	fmt.Println(strings.Repeat("-", 80))

	rows := make([]Row, 0, len(resolvers))

	for _, r := range resolvers {
		samples := make([]Sample, 0, *count)
		for i := 0; i < *count; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			qname := *domain
			if *cold {
				qname = randomLabel() + "." + *domain
			}
			start := time.Now()
			err := lookup(ctx, r.Addr, qname, *network)
			d := time.Since(start)
			cancel()

			samples = append(samples, Sample{Duration: d, Err: err})
		}
		stats := summarize(samples)
		rows = append(rows, Row{Name: r.Name, Stats: stats, Samples: samples})
	}

	printTable(rows)

	if *outCSV != "" {
		if err := writeCSV(*outCSV, rows); err != nil {
			fmt.Fprintf(os.Stderr, "CSV write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nCSV written to: %s\n", *outCSV)
	}
}

func parseResolvers(s string) []ResolverCfg {
	parts := strings.Split(s, ",")
	var out []ResolverCfg
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		addr := strings.TrimSpace(kv[1])
		out = append(out, ResolverCfg{Name: name, Addr: addr})
	}
	return out
}

// lookup performs a single A/AAAA lookup against a specific resolver using net.Resolver.
func lookup(ctx context.Context, resolverAddr, name, network string) error {
	host, port, ok := strings.Cut(resolverAddr, ":")
	if !ok {
		host = resolverAddr
		port = "53"
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, n, address string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(host, port))
		},
	}

	switch strings.ToLower(network) {
	case "ip4", "ipv4":
		_, err := r.LookupIP(ctx, "ip4", name)
		return err
	case "ip6", "ipv6":
		_, err := r.LookupIP(ctx, "ip6", name)
		return err
	default:
		_, err := r.LookupIP(ctx, "ip4", name)
		return err
	}
}

func summarize(samples []Sample) Stats {
	var stats Stats
	stats.Count = len(samples)
	stats.Min = time.Duration(math.MaxInt64)
	for _, s := range samples {
		if s.Err == nil {
			stats.Successes++
			if s.Duration < stats.Min {
				stats.Min = s.Duration
			}
			if s.Duration > stats.Max {
				stats.Max = s.Duration
			}
			stats.DurationsMs = append(stats.DurationsMs, float64(s.Duration.Microseconds())/1000.0)
		} else {
			stats.Errors = append(stats.Errors, s.Err)
		}
	}
	if stats.Successes == 0 {
		stats.Min = 0
		stats.Max = 0
		return stats
	}
	// avg
	var sum float64
	for _, v := range stats.DurationsMs {
		sum += v
	}
	avgMs := sum / float64(stats.Successes)
	stats.Avg = time.Duration(avgMs * float64(time.Millisecond))

	// median & p95
	ms := make([]float64, len(stats.DurationsMs))
	copy(ms, stats.DurationsMs)
	sort.Float64s(ms)
	stats.Median = time.Duration(percentile(ms, 50) * float64(time.Millisecond))
	stats.P95 = time.Duration(percentile(ms, 95) * float64(time.Millisecond))

	return stats
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	pos := (p / 100) * float64(len(sorted)-1)
	l := int(math.Floor(pos))
	u := int(math.Ceil(pos))
	if l == u {
		return sorted[l]
	}
	frac := pos - float64(l)
	return sorted[l]*(1-frac) + sorted[u]*frac
}

func printTable(rows []Row) {
	fmt.Printf("%-12s  %6s  %6s  %6s  %6s  %6s  %9s\n",
		"Resolver", "Min", "Avg", "Med", "p95", "Max", "Success%")
	fmt.Println(strings.Repeat("-", 72))

	for _, r := range rows {
		s := r.Stats
		successPct := 0.0
		if s.Count > 0 {
			successPct = 100.0 * float64(s.Successes) / float64(s.Count)
		}
		fmt.Printf("%-12s  %6s  %6s  %6s  %6s  %6s  %8.1f%%\n",
			r.Name,
			durFmt(s.Min),
			durFmt(s.Avg),
			durFmt(s.Median),
			durFmt(s.P95),
			durFmt(s.Max),
			successPct,
		)
		if len(s.Errors) > 0 {
			uniq := uniqueErrors(s.Errors)
			for _, e := range uniq {
				fmt.Printf("  ! %s\n", e)
			}
		}
	}
}

func writeCSV(path string, rows []Row) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"resolver", "count", "successes", "min_ms", "avg_ms", "median_ms", "p95_ms", "max_ms", "errors"}); err != nil {
		return err
	}
	for _, r := range rows {
		s := r.Stats
		errStr := ""
		if len(s.Errors) > 0 {
			errStr = strings.Join(errorStrings(uniqueErrors(s.Errors)), " | ")
		}
		row := []string{
			r.Name,
			fmt.Sprintf("%d", s.Count),
			fmt.Sprintf("%d", s.Successes),
			fmt.Sprintf("%.3f", float64(s.Min.Microseconds())/1000.0),
			fmt.Sprintf("%.3f", float64(s.Avg.Microseconds())/1000.0),
			fmt.Sprintf("%.3f", float64(s.Median.Microseconds())/1000.0),
			fmt.Sprintf("%.3f", float64(s.P95.Microseconds())/1000.0),
			fmt.Sprintf("%.3f", float64(s.Max.Microseconds())/1000.0),
			errStr,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	if err := w.Write([]string{}); err != nil {
		return err
	}
	if err := w.Write([]string{"resolver", "run_index", "duration_ms", "error"}); err != nil {
		return err
	}
	for _, r := range rows {
		for i, s := range r.Samples {
			errStr := ""
			if s.Err != nil {
				errStr = s.Err.Error()
			}
			row := []string{
				r.Name,
				fmt.Sprintf("%d", i),
				fmt.Sprintf("%.3f", float64(s.Duration.Microseconds())/1000.0),
				errStr,
			}
			if err := w.Write(row); err != nil {
				return err
			}
		}
	}
	return nil
}

func durFmt(d time.Duration) string {
	if d <= 0 {
		return "--"
	}
	ms := float64(d.Microseconds()) / 1000.0
	return fmt.Sprintf("%.1fms", ms)
}

func uniqueErrors(errs []error) []error {
	seen := make(map[string]bool)
	var out []error
	for _, e := range errs {
		if e == nil {
			continue
		}
		if !seen[e.Error()] {
			seen[e.Error()] = true
			out = append(out, e)
		}
	}
	return out
}

func errorStrings(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out
}

func randomLabel() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	hex := make([]byte, len(b)*2)
	const hexdigits = "0123456789abcdef"
	for i, v := range b {
		hex[i*2] = hexdigits[v>>4]
		hex[i*2+1] = hexdigits[v&0x0f]
	}
	return string(hex)
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}


