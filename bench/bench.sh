#!/usr/bin/env bash
# ./bench/bench.sh [runs]
#
# A deliberately small benchmark: chalk showing one tab (start, read
# bench/tab.json, parse the embedded menu, print, exit), built three ways and
# run under Docker. Cold start is the median of [runs] (default 100) separate
# processes, after five that don't count. Writes bench/results.md.
#
# Plain bash 3.2, as macOS ships it: no associative arrays.
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT=$PWD
N=${1:-100}
ARCH=$(docker version --format '{{.Server.Arch}}')
OUT=bench/bin/results
mkdir -p "$OUT"

label() {
  case $1 in
    go) echo "Go" ;;
    jit) echo ".NET, JIT" ;;
    aot) echo ".NET, Native AOT" ;;
  esac
}

(cd bench/runner && GOOS=linux GOARCH="$ARCH" CGO_ENABLED=0 go build -ldflags="-s -w" -o "../bin/runner-linux-$ARCH" .)

# Cold compile times, in throwaway containers with empty build caches. The
# .NET packages are restored first and not timed, since that is the network.
echo "== compile times"
docker run --rm -v "$ROOT:/repo:ro" golang:1.27.3 bash -c '
  set -e
  cp -r /repo/go /w && cd /w
  s=$(date +%s%N); CGO_ENABLED=0 go build -o /tmp/chalk ./cmd/chalk; e=$(date +%s%N)
  echo $(( (e - s) / 1000000 ))' > "$OUT/go.build"
dotnet_build() { # $1 extra publish args, $2 setup
  docker run --rm -v "$ROOT:/repo:ro" mcr.microsoft.com/dotnet/sdk:10.0 bash -c "
    set -e
    $2
    mkdir /w && cp -r /repo/contract /repo/dotnet /w/ && cd /w && rm dotnet/global.json
    dotnet restore dotnet/src/Tabanco.Cli $1 >/dev/null
    s=\$(date +%s%N); dotnet publish dotnet/src/Tabanco.Cli -c Release --no-restore $1 -o /tmp/out >/dev/null; e=\$(date +%s%N)
    echo \$(( (e - s) / 1000000 ))"
}
dotnet_build "" "" > "$OUT/jit.build"
dotnet_build "-p:PublishAot=true" \
  "apt-get update -qq >/dev/null && apt-get install -y -qq --no-install-recommends clang zlib1g-dev >/dev/null" > "$OUT/aot.build"

for v in go jit aot; do
  echo "== $v: building image"
  docker build -q -f "bench/$v.Dockerfile" -t "tabanco-bench-$v" . >/dev/null
  echo "== $v: $N runs"
  docker run --rm \
    -v "$ROOT/bench/bin/runner-linux-$ARCH:/runner:ro" \
    -v "$ROOT/bench/tab.json:/tab.json:ro" \
    --entrypoint /runner "tabanco-bench-$v" "$N" /app/chalk --file /tab.json show Mari > "$OUT/$v.json"
  # What the app adds to its base image: the binary, or the published folder.
  docker run --rm --entrypoint du "tabanco-bench-$v" -sk /app | cut -f1 > "$OUT/$v.ships"
  docker image inspect "tabanco-bench-$v" --format '{{.Size}}' > "$OUT/$v.image"
  label "$v" > "$OUT/$v.label"
done

python3 - "$N" "$ARCH" "$OUT" > bench/results.md <<'PY'
import json, sys, pathlib
n, arch, out = sys.argv[1], sys.argv[2], pathlib.Path(sys.argv[3])
read = lambda v, ext: (out / f"{v}.{ext}").read_text().strip()
rows = {v: (read(v, "label"), json.loads(read(v, "json")), int(read(v, "ships")), int(read(v, "image")), int(read(v, "build")))
        for v in ("go", "jit", "aot")}
outs = {r[1]["stdout"] for r in rows.values()}
assert len(outs) == 1, "the three builds printed different tabs"
print(f"chalk showing one tab, {n} cold starts each under Docker on linux/{arch}.\n")
print("| Build | Cold start (median) | Fastest | 90th percentile | Peak memory | App | Image | Compile |")
print("|---|---|---|---|---|---|---|---|")
for label, r, ships_kb, image_b, build_ms in rows.values():
    print(f"| {label} | **{r['median_ms']:.2f} ms** | {r['min_ms']:.2f} ms | {r['p90_ms']:.2f} ms "
          f"| {r['median_rss_kb']/1024:.1f} MB | {ships_kb/1024:.1f} MB | {image_b/1e6:.0f} MB | {build_ms/1000:.1f} s |")
print("\nAll three printed the same tab:\n\n```\n" + outs.pop() + "```")
PY
cat bench/results.md
