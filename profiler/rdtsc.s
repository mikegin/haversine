// https://github.com/efficient/qlease/blob/f965b871ecffa5cd7aa8ac1ad9bf8fdaf8d90f99/src/rdtsc/rdtsc.s
// func Cputicks(void) (n uint64)
TEXT ·ReadCPUTimer(SB),7,$0
    RDTSC
    SHLQ  $32, DX
    ADDQ  DX, AX
    MOVQ  AX, n+0(FP)
    RET
