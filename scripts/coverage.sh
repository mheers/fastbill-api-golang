#!/usr/bin/env bash
# Runs the test suite with coverage and fails unless every library package
# reaches the threshold. examples/ and internal/fastbilltest are excluded.
set -euo pipefail

threshold="${1:-90}"
packages=$(go list ./... | grep -vE '/(examples|internal/fastbilltest)($|/)')

go test -race -covermode=atomic -coverprofile=coverage.out ${packages}

awk -v threshold="${threshold}" '
	NR == 1 { next }
	{
		split($1, location, ":")
		file = location[1]
		n = split(file, parts, "/")
		pkg = ""
		for (i = 1; i < n; i++) {
			pkg = pkg (i > 1 ? "/" : "") parts[i]
		}
		total[pkg] += $2
		if ($3 + 0 > 0) {
			covered[pkg] += $2
		}
		totalAll += $2
		if ($3 + 0 > 0) {
			coveredAll += $2
		}
	}
	END {
		failed = 0
		for (pkg in total) {
			pct = 100 * covered[pkg] / total[pkg]
			printf "%-78s %6.1f%%\n", pkg, pct
			if (pct + 0 < threshold + 0) {
				printf "FAIL: %s is %.1f%%, below %.0f%%\n", pkg, pct, threshold > "/dev/stderr"
				failed = 1
			}
		}
		pct = 100 * coveredAll / totalAll
		printf "%-78s %6.1f%%\n", "TOTAL", pct
		if (pct + 0 < threshold + 0) {
			printf "FAIL: total is %.1f%%, below %.0f%%\n", pct, threshold > "/dev/stderr"
			failed = 1
		}
		exit failed
	}
' coverage.out
