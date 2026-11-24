chalk showing one tab, 100 cold starts each under Docker on linux/arm64.

| Build | Cold start (median) | Fastest | 90th percentile | Peak memory | App | Image | Compile |
|---|---|---|---|---|---|---|---|
| Go | **0.88 ms** | 0.72 ms | 1.08 ms | 3.3 MB | 3.7 MB | 48 MB | 1.6 s |
| .NET, JIT | **24.33 ms** | 22.98 ms | 29.19 ms | 34.5 MB | 0.2 MB | 81 MB | 1.2 s |
| .NET, Native AOT | **2.13 ms** | 1.86 ms | 2.41 ms | 9.1 MB | 2.6 MB | 47 MB | 2.8 s |

All three printed the same tab:

```
Mari
  Fino           ||||/ |            6    12,00 €
  Manzanilla     ||                 2     4,00 €
  Pedro Ximénez  |                  1     2,50 €
  Chicharrones   |                  1     4,00 €
  Payoyo         |                  1     4,50 €
  ----------------------------------------------
  total                                  27,00 €
```
