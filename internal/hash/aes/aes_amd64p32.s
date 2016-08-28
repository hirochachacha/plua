TEXT	·cpuid(SB),4,$0
	MOVQ	$0, AX
	CPUID
	CMPQ	AX, $0
	JE	nocpuinfo
	MOVQ	$1, AX
	CPUID
	MOVL	CX, ·cpuid_ecx(SB)
nocpuinfo:
	RET

// hash function using AES hardware instructions
// For now, our one amd64p32 system (NaCl) does not
// support using AES instructions, so have not bothered to
// write the implementations. Can copy and adjust the ones
// in asm_amd64.s when the time comes.

TEXT ·aeshash(SB),4,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT ·aeshash32(SB),4,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT ·aeshash64(SB),4,$0-20
	MOVL	AX, ret+16(FP)
	RET
