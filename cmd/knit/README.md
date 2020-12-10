knit usage examples
===================

All examples captured on a 4 cores (with HT, thus 2 physical cores) laptop.


checking CPU affinities of all processes. Here we grep `knit` for the sake of the brevity, the output
would be very verbose.
```bash
knit cpuaff | grep knit
PID 166574 (knit                            ) TID 166574 (knit            ) can run on [0 1 2 3]
PID 166574 (knit                            ) TID 166576 (knit            ) can run on [0 1 2 3]
PID 166574 (knit                            ) TID 166577 (knit            ) can run on [0 1 2 3]
PID 166574 (knit                            ) TID 166578 (knit            ) can run on [0 1 2 3]
PID 166574 (knit                            ) TID 166579 (knit            ) can run on [0 1 2 3]
PID 166574 (knit                            ) TID 166580 (knit            ) can run on [0 1 2 3]
```

Checking which processes can run on CPU #3. There is no affinity constraint, all process can run there.
Again we grep `lnit` for the sake of brevity.
```bash
$ ./_output/knit cpuaff -C 3 | grep knit
PID 166617 (knit                            ) TID 166617 (knit            ) can run on [3]
PID 166617 (knit                            ) TID 166619 (knit            ) can run on [3]
PID 166617 (knit                            ) TID 166620 (knit            ) can run on [3]
PID 166617 (knit                            ) TID 166621 (knit            ) can run on [3]
PID 166617 (knit                            ) TID 166622 (knit            ) can run on [3]
PID 166617 (knit                            ) TID 166623 (knit            ) can run on [3]

```

Checking all the IRQ affinities.
```bash
$ knit irqaff
IRQ   0 [                        ]: can run on [0 1 2 3]
IRQ   1 [                   i8042]: can run on [0 1 2 3]
IRQ   2 [                        ]: can run on [0 1 2 3]
IRQ   3 [                        ]: can run on [0 1 2 3]
IRQ   4 [                        ]: can run on [0 1 2 3]
IRQ   5 [                        ]: can run on [0 1 2 3]
IRQ   6 [                        ]: can run on [0 1 2 3]
IRQ   7 [                        ]: can run on [0 1 2 3]
IRQ   8 [                    rtc0]: can run on [0 1 2 3]
IRQ   9 [                    acpi]: can run on [0 1 2 3]
IRQ  10 [                        ]: can run on [0 1 2 3]
IRQ  11 [                        ]: can run on [0 1 2 3]
IRQ  12 [                   i8042]: can run on [0 1 2 3]
IRQ  13 [                        ]: can run on [0 1 2 3]
IRQ  14 [                        ]: can run on [0 1 2 3]
IRQ  15 [                        ]: can run on [0 1 2 3]
IRQ  16 [              i801_smbus]: can run on [0 1 2 3]
IRQ 120 [                        ]: can run on [0]
IRQ 121 [                        ]: can run on [0]
IRQ 125 [                xhci_hcd]: can run on [0 1 2 3]
IRQ 126 [                 nvme0q0]: can run on [0 1 2 3]
IRQ 127 [               enp0s31f6]: can run on [0 1 2 3]
IRQ 128 [                 nvme0q1]: can run on [0]
IRQ 129 [                 nvme0q2]: can run on [1]
IRQ 130 [                 nvme0q3]: can run on [2]
IRQ 131 [                 nvme0q4]: can run on [3]
IRQ 132 [                    i915]: can run on [0 1 2 3]
IRQ 133 [                  mei_me]: can run on [0 1 2 3]
IRQ 134 [                 iwlwifi]: can run on [0 1 2 3]
IRQ 135 [     snd_hda_intel:card0]: can run on [0 1 2 3]
IRQ 136 [              rmi4_smbus]: can run on [0 1 2 3]
IRQ 137 [            rmi4-00.fn34]: can run on [0 1 2 3]
IRQ 138 [            rmi4-00.fn01]: can run on [0 1 2 3]
IRQ 139 [            rmi4-00.fn03]: can run on [0 1 2 3]
IRQ 140 [            rmi4-00.fn11]: can run on [0 1 2 3]
IRQ 141 [            rmi4-00.fn11]: can run on [0 1 2 3]
IRQ 142 [            rmi4-00.fn30]: can run on [0 1 2 3]
```

Checking which IRQ will be served on CPUs #1 and #2
```bash
$ knit irqaff -C 1,2
IRQ   0 [                        ]: can run on [1 2]
IRQ   1 [                   i8042]: can run on [1 2]
IRQ   2 [                        ]: can run on [1 2]
IRQ   3 [                        ]: can run on [1 2]
IRQ   4 [                        ]: can run on [1 2]
IRQ   5 [                        ]: can run on [1 2]
IRQ   6 [                        ]: can run on [1 2]
IRQ   7 [                        ]: can run on [1 2]
IRQ   8 [                    rtc0]: can run on [1 2]
IRQ   9 [                    acpi]: can run on [1 2]
IRQ  10 [                        ]: can run on [1 2]
IRQ  11 [                        ]: can run on [1 2]
IRQ  12 [                   i8042]: can run on [1 2]
IRQ  13 [                        ]: can run on [1 2]
IRQ  14 [                        ]: can run on [1 2]
IRQ  15 [                        ]: can run on [1 2]
IRQ  16 [              i801_smbus]: can run on [1 2]
IRQ 125 [                xhci_hcd]: can run on [1 2]
IRQ 126 [                 nvme0q0]: can run on [1 2]
IRQ 127 [               enp0s31f6]: can run on [1 2]
IRQ 129 [                 nvme0q2]: can run on [1]
IRQ 130 [                 nvme0q3]: can run on [2]
IRQ 132 [                    i915]: can run on [1 2]
IRQ 133 [                  mei_me]: can run on [1 2]
IRQ 134 [                 iwlwifi]: can run on [1 2]
IRQ 135 [     snd_hda_intel:card0]: can run on [1 2]
IRQ 136 [              rmi4_smbus]: can run on [1 2]
IRQ 137 [            rmi4-00.fn34]: can run on [1 2]
IRQ 138 [            rmi4-00.fn01]: can run on [1 2]
IRQ 139 [            rmi4-00.fn03]: can run on [1 2]
IRQ 140 [            rmi4-00.fn11]: can run on [1 2]
IRQ 141 [            rmi4-00.fn11]: can run on [1 2]
IRQ 142 [            rmi4-00.fn30]: can run on [1 2]
```

Checking softirqs affinity. All CPUs served softirqs.
```bash
$ knit irqaff -s
      HI = 0-3
   TIMER = 0-3
  NET_TX = 0-3
  NET_RX = 0-3
   BLOCK = 0-3
IRQ_POLL = 
 TASKLET = 0-3
   SCHED = 0-3
 HRTIMER = 0-3
     RCU = 0-3
```

Checking if CPUs #1 and #3 served softirqs. They did.
```bash
$ ./_output/knit irqaff -s -C 1,3
      HI = 1,3
   TIMER = 1,3
  NET_TX = 1,3
  NET_RX = 1,3
   BLOCK = 1,3
IRQ_POLL = 
 TASKLET = 1,3
   SCHED = 1,3
 HRTIMER = 1,3
     RCU = 1,3
```
