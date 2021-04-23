## cpulist

```bash
$ cpulist  # output varies depending on the machine
0
1
2
3
$ cpulist -c '0,2-4,7'
0
2
3
4
7
$ for CPU in $( cpulist -c '0,2-4,7' ); do echo "considering CPU=${CPU}"; done
considering CPU=0
considering CPU=2
considering CPU=3
considering CPU=4
considering CPU=7
$ echo "0-4" | ./_output/cpulist -f -
0
1
2
3
4
```
