package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MyExporter struct {
	Info  interface{}
	Error string
}

var (
	domainExpiration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_expiration_time",
			Help: "Time remaining until a domain expires",
		},
		[]string{"domain"},
	)
)

func init() {
	prometheus.MustRegister(domainExpiration)
}

// 基于 context 的缓存方式
func getDomainExpiration(ctx context.Context, domain string, cache map[string]float64) (float64, error) {
	// Check if the expiration time for this domain is already in the cache
	cache = make(map[string]float64)

	if exp, ok := cache[domain]; ok {
		return exp, nil
	}

	// 创建一个超时定时器
	timer := time.NewTimer(5 * time.Second)

	// 创建一个结果通道
	result := make(chan *MyExporter, 1)

	// 在一个单独的goroutine中执行WHOIS查询
	go func() {
		info, err := whois.Whois(domain)
		if err != nil {
			result <- &MyExporter{
				Error: err.Error(),
			}
			return
		}
		result <- &MyExporter{
			Info: info,
		}
	}()

	// 选择两个结果：一个是从定时器中获取的结果，另一个是从通道中获取的结果
	select {
	case <-timer.C:
		return 0, errors.New("get whois info timeout")
	case res := <-result:
		if res.Error != "" {
			return 0, errors.New(res.Error)
		}

		parsed, err := whoisparser.Parse(res.Info.(string))
		if err != nil {
			return 0, err
		}

		// expiration, err := time.Parse("2006-01-02 15:04:05", parsed.Domain.ExpirationDate)
		// 先把不同格式的时间转换成统一的格式，只保留年-月-日，得到字符串
		expiration, err := formatTime(parsed.Domain.ExpirationDate)
		if err != nil {
			return 0, fmt.Errorf("error parsing expiration date: %v", err)
		}

		// 再把字符串转成时间类型
		t1, err := time.Parse("2006-01-02", expiration)
		if err != nil {
			// handle error
			return 0, fmt.Errorf("error parsing expiration to date: %v", err)
		}
		timeRemaining := t1.Sub(time.Now()).Hours() / 24
		cache[domain] = timeRemaining
		return timeRemaining, nil
	case <-ctx.Done():
		return cache[domain], ctx.Err()
	}
}

func collectMetrics() {
	domains := []string{
		"baidu.com",
		"jd.com",
		"meitu.com",
	}

	// 基于 context 的缓存方式
	var cache map[string]float64
	for _, domain := range domains {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		timeRemaining, err := getDomainExpiration(ctx, domain, cache)
		if err != nil {
			continue
		}
		domainExpiration.WithLabelValues(domain).Set(timeRemaining)
		cancel()
	}
}

func formatTime(timestring string) (string, error) {
	var layout string
	if len(timestring) == 20 {
		layout = "2006-01-02T15:04:05Z"
	} else {
		layout = "2006-01-02 15:04:05"
	}
	t, err := time.Parse(layout, timestring)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

func main() {
	go func() {
		for {
			collectMetrics()
			// cache metrics for 10 minutes
			time.Sleep(10 * time.Minute)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
