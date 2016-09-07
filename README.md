README
------

[![Build Status](https://drone.io/bitbucket.org/weberc2/fastcsv/status.png)](https://drone.io/bitbucket.org/weberc2/fastcsv/latest)

**Testing CI** Remove this!

**Warning**: This library is still in alpha. API and behavior are expected to
change.

Fast CSV reader library. Reads CSV files 5X faster than `encoding/csv`.

```
BenchmarkStdCsv-4   100   140391400 ns/op   15258771 B/op  696086 allocs/op
BenchmarkFastCsv-4  500    25303160 ns/op       2224 B/op       4 allocs/op
```

## Differences from `encoding/csv`

* This library is fast because it only handles standard delimiters (`,` and
`"`) instead of the full gamut of UTF-8 runes accepted by the standard library
implementation.
* This implementation provides a streaming interface; only one row is valid at
a time; subsequent calls to `Read()` invalidate the previously read row. This
keeps allocs down to `O(1)` instead of `O(N)`.
* This implementation deals in `[]byte` instead of `string`. This also helps to
keep allocs down to `O(1)` instead of `O(N)`.
