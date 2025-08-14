# 📊 Performance Profiling Guide

This guide covers the comprehensive performance profiling and monitoring capabilities built into Template Architecture Lint.

## 🎯 Overview

The application includes two complementary profiling systems:
- **🔬 pprof Integration** - Low-level runtime profiling via Go's built-in pprof
- **📈 Performance Endpoints** - High-level application metrics and health monitoring

## 🚀 Quick Start

### 1. Enable Development Mode
Performance profiling is available in development and debug environments:

```bash
# Set environment to enable profiling
export APP_ENVIRONMENT=development

# Start the server
just run
# or
go run cmd/server/main.go
```

### 2. Verify Profiling is Active
Check that profiling endpoints are available:

```bash
# Check pprof availability
curl http://localhost:8080/debug/metrics

# Check performance monitoring
curl http://localhost:8080/performance/stats
```

## 🔬 pprof Integration

### Available Endpoints

| Endpoint | Description | Usage |
|----------|-------------|-------|
| `/debug/pprof/` | Profile index page | Browser access to all profiles |
| `/debug/pprof/profile` | CPU profiling | `?seconds=30` for 30-second sample |
| `/debug/pprof/heap` | Memory heap analysis | Current heap allocations |
| `/debug/pprof/goroutine` | Goroutine dump | All active goroutines |
| `/debug/pprof/allocs` | Memory allocation history | All allocations since start |
| `/debug/pprof/block` | Blocking profile | Blocking operations |
| `/debug/pprof/mutex` | Mutex contention | Lock contention analysis |
| `/debug/pprof/trace` | Execution trace | `?seconds=10` for trace capture |

### CPU Profiling

**Capture a 30-second CPU profile:**
```bash
curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o cpu.prof
```

**Analyze with go tool pprof:**
```bash
# Interactive analysis
go tool pprof cpu.prof

# Commands within pprof:
(pprof) top10        # Top 10 functions by CPU usage
(pprof) list main    # Source code view of main function
(pprof) web          # Generate SVG visualization (requires graphviz)
(pprof) png          # Generate PNG visualization

# Direct web interface
go tool pprof -http=:8081 cpu.prof
# Visit http://localhost:8081 for interactive web UI
```

### Memory Profiling

**Capture heap profile:**
```bash
curl http://localhost:8080/debug/pprof/heap -o heap.prof
```

**Analyze memory usage:**
```bash
go tool pprof heap.prof

# Common pprof commands for memory:
(pprof) top10 -cum   # Top by cumulative allocation
(pprof) top10 inuse  # Top by current memory usage
(pprof) list NewUser # Memory usage in specific function
(pprof) svg          # Generate memory usage visualization
```

**Track allocation history:**
```bash
curl http://localhost:8080/debug/pprof/allocs -o allocs.prof
go tool pprof allocs.prof
```

### Goroutine Analysis

**Capture goroutine dump:**
```bash
curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof
```

**Analyze concurrency:**
```bash
go tool pprof goroutine.prof

# Goroutine analysis commands:
(pprof) top          # Functions creating most goroutines
(pprof) traces       # Stack traces of all goroutines
(pprof) web          # Visualization of goroutine relationships
```

### Blocking and Mutex Profiling

**Enable blocking profile (add to main.go for detailed analysis):**
```go
import _ "runtime/debug"

func init() {
    runtime.SetBlockProfileRate(1)
    runtime.SetMutexProfileFraction(1)
}
```

**Capture and analyze:**
```bash
# Blocking operations
curl http://localhost:8080/debug/pprof/block -o block.prof
go tool pprof block.prof

# Mutex contention
curl http://localhost:8080/debug/pprof/mutex -o mutex.prof
go tool pprof mutex.prof
```

### Execution Tracing

**Capture execution trace:**
```bash
curl "http://localhost:8080/debug/pprof/trace?seconds=10" -o trace.out
```

**Analyze with trace tool:**
```bash
go tool trace trace.out
# Opens web browser with detailed execution timeline
```

## 📈 Performance Monitoring Endpoints

### Runtime Statistics
**Endpoint:** `GET /performance/stats`

Provides comprehensive runtime metrics:
```bash
curl http://localhost:8080/performance/stats | jq
```

**Response includes:**
- Memory usage (heap, stack, allocations)
- Garbage collection statistics
- Goroutine counts
- CPU and system information
- Application uptime

### Health Metrics
**Endpoint:** `GET /performance/health`

Application health assessment with status codes:
```bash
curl http://localhost:8080/performance/health
```

**Status Codes:**
- `200` - Healthy
- `202` - Warning (acceptable performance)
- `503` - Critical (performance issues detected)

**Health Thresholds:**
- Max Goroutines: 1,000
- Max Heap Usage: 512 MB
- Max GC Pause: 100ms (warning), 500ms (critical)

### Force Garbage Collection
**Endpoint:** `POST /performance/gc` *(Development only)*

Manually trigger GC and measure impact:
```bash
curl -X POST http://localhost:8080/performance/gc
```

**Response includes:**
- Memory freed
- Objects collected
- GC duration

### Memory Dump
**Endpoint:** `GET /performance/memory` *(Development only)*

Detailed memory allocation breakdown:
```bash
curl http://localhost:8080/performance/memory | jq
```

### Debug Information
**Endpoint:** `GET /performance/info`

Build and runtime information:
```bash
curl http://localhost:8080/performance/info | jq
```

## 🛠️ Development Workflow

### 1. Load Testing Setup

**Create load test script:**
```bash
#!/bin/bash
# load_test.sh

echo "Starting load test..."

# Background processes hitting different endpoints
for i in {1..10}; do
    curl -s http://localhost:8080/api/v1/users > /dev/null &
    curl -s http://localhost:8080/users > /dev/null &
done

echo "Load test running... capture profiles now"
sleep 60
echo "Load test complete"
```

**Profile under load:**
```bash
# Terminal 1: Start load test
./load_test.sh

# Terminal 2: Capture profiles during load
curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o load_cpu.prof
curl http://localhost:8080/debug/pprof/heap -o load_heap.prof
curl http://localhost:8080/performance/stats > load_stats.json
```

### 2. Performance Regression Detection

**Baseline establishment:**
```bash
# Create baseline profiles
mkdir profiles/baseline
curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o profiles/baseline/cpu.prof
curl http://localhost:8080/debug/pprof/heap -o profiles/baseline/heap.prof
curl http://localhost:8080/performance/stats > profiles/baseline/stats.json
```

**Regression testing:**
```bash
# After code changes
mkdir profiles/current
curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o profiles/current/cpu.prof
curl http://localhost:8080/debug/pprof/heap -o profiles/current/heap.prof

# Compare profiles
go tool pprof -base profiles/baseline/cpu.prof profiles/current/cpu.prof
```

### 3. CI/CD Integration

**Add to GitHub Actions:**
```yaml
- name: Performance Baseline Check
  run: |
    # Start server in background
    just run &
    SERVER_PID=$!
    sleep 5

    # Capture performance metrics
    curl http://localhost:8080/performance/health
    curl http://localhost:8080/performance/stats > perf_stats.json

    # Cleanup
    kill $SERVER_PID

    # Fail if critical performance issues detected
    if grep -q '"status":"critical"' perf_stats.json; then
      echo "Critical performance issues detected"
      exit 1
    fi
```

## 🔧 Justfile Integration

The following commands are available for profiling:

```bash
# Performance profiling
just profile-cpu           # Capture 30-second CPU profile
just profile-heap          # Capture heap profile
just profile-goroutines    # Capture goroutine dump
just profile-trace         # Capture 10-second execution trace

# Analysis
just analyze-cpu           # Open CPU profile in browser
just analyze-heap          # Open heap profile in browser
just compare-profiles      # Compare current with baseline
```

**Add to justfile:**
```justfile
# Capture CPU profile for 30 seconds
profile-cpu:
    @echo "📊 Capturing CPU profile (30 seconds)..."
    curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o cpu.prof
    @echo "✅ CPU profile saved to cpu.prof"

# Capture heap profile
profile-heap:
    @echo "📊 Capturing heap profile..."
    curl http://localhost:8080/debug/pprof/heap -o heap.prof
    @echo "✅ Heap profile saved to heap.prof"

# Analyze CPU profile in browser
analyze-cpu:
    @echo "🔍 Opening CPU profile analysis..."
    go tool pprof -http=:8081 cpu.prof

# Analyze heap profile in browser  
analyze-heap:
    @echo "🔍 Opening heap profile analysis..."
    go tool pprof -http=:8081 heap.prof
```

## 🎯 Common Performance Issues

### High Memory Usage
```bash
# Investigate heap allocations
curl http://localhost:8080/debug/pprof/heap -o heap.prof
go tool pprof heap.prof

# Check for memory leaks
(pprof) top10 inuse    # Current memory usage
(pprof) top10 alloc    # Total allocations
(pprof) list problematicFunction
```

### High CPU Usage
```bash
# Profile CPU usage
curl "http://localhost:8080/debug/pprof/profile?seconds=30" -o cpu.prof
go tool pprof cpu.prof

# Find hot spots
(pprof) top10          # Top CPU consumers
(pprof) web            # Visual call graph
```

### Goroutine Leaks
```bash
# Check goroutine count
curl http://localhost:8080/performance/stats | jq '.goroutines'

# Analyze goroutines
curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof
go tool pprof goroutine.prof
(pprof) traces         # See all goroutine stack traces
```

### GC Pressure
```bash
# Monitor GC behavior
curl http://localhost:8080/performance/stats | jq '.gc'

# Force GC and measure impact
curl -X POST http://localhost:8080/performance/gc

# Optimize allocations using heap profile
curl http://localhost:8080/debug/pprof/allocs -o allocs.prof
go tool pprof allocs.prof
```

## 📋 Best Practices

### 1. **Regular Monitoring**
- Set up automated health checks
- Monitor key metrics (goroutines, heap size, GC pause times)
- Establish performance baselines for regression detection

### 2. **Profiling Strategy**
- Profile in production-like environments
- Use representative workloads
- Profile different scenarios (startup, steady state, peak load)
- Compare profiles before and after changes

### 3. **Security Considerations**
- Profiling endpoints are disabled in production
- Use environment variables to control profiling access
- Never expose profiling endpoints publicly
- Monitor profiling endpoint access in logs

### 4. **Performance Optimization**
- Focus on the biggest performance bottlenecks first
- Use allocation profiles to reduce GC pressure
- Optimize hot paths identified by CPU profiling
- Monitor goroutine counts to prevent leaks

## 🔒 Security Notes

- **Development Only**: pprof endpoints only available in development/debug mode
- **No Production Access**: Profiling disabled in production environments
- **Local Access**: Endpoints bound to localhost in secure deployments
- **Authentication**: Consider adding authentication for sensitive profiling data

## 📚 Additional Resources

- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Go Execution Tracer](https://pkg.go.dev/runtime/trace)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)
