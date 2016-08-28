TEXT	·cpuid(SB),4,$0
	RET

// AES hashing not implemented for ARM
TEXT ·aeshash(SB),4,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT ·aeshash32(SB),4,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT ·aeshash64(SB),4,$-4-0
	MOVW	$0, R0
	MOVW	(R0), R1
