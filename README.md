# libinjection
libinjection is a Go porting of the libinjection([http://www.client9.com/projects/libinjection/](http://www.client9.com/projects/libinjection/)).

## How to use
### SQLi Example
```go
result, fingerprint := libinjection.IsSQLi("' OR '1'='1' --")
```

### XSS Example
```go

```

## Benchmark
### SQLi benchmark
```go

```

### XSS benchmark
```go

```

## License
libinjection-golang is distributed under the same license as libinjection.