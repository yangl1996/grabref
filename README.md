# Grabref

Grab references from manuscript. Specifically it looks for patterns like
`($YEAR)` (such as `(Alice, Bob and Charlie 2018)` and
`Alice et al. (2018)`) and makes a list.

## Usage

```
Usage of ./grabref:
  -dedup
    	deduplicate references, implies sorted
  -etal
    	add 'et al.' to author list if it appears in input
  -input string
    	path to the text file, use stdin if left empty
  -output string
    	path to the output file, stdout if left empty
  -sorted
    	sort authors lexicographically
```

