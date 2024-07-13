# Haversine distance data generation, parsing and profiling

From Casey Muratori's Performance Aware Programming Series:
https://www.computerenhance.com/p/table-of-contents

### generate
```
go run generate_haversine/main.go
```

```
go run generate_haversine/main.go uniform 12312316245 10000000
```

### parse
```
 go run . generate_haversine/data_10000000_pairs.json generate_haversine/data_10000000_haveranswers.f64
```
### parse and profile
```
PROFILER=true go run . generate_haversine/data_10000000_pairs.json generate_haversine data_10000000_haveranswers.f64
```



