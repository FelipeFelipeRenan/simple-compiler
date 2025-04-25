	.text
	.file	"testeeee-157822942.ll"
	.globl	sum                             # -- Begin function sum
	.p2align	4, 0x90
	.type	sum,@function
sum:                                    # @sum
	.cfi_startproc
# %bb.0:                                # %entry
                                        # kill: def $esi killed $esi def $rsi
                                        # kill: def $edi killed $edi def $rdi
	movl	%edi, -4(%rsp)
	movl	%esi, -8(%rsp)
	leal	(%rdi,%rsi), %eax
	retq
.Lfunc_end0:
	.size	sum, .Lfunc_end0-sum
	.cfi_endproc
                                        # -- End function
	.globl	main                            # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %entry
	subq	$24, %rsp
	.cfi_def_cfa_offset 32
	movl	$2, %edi
	movl	$-10, %esi
	callq	sum@PLT
	movl	%eax, 12(%rsp)
	movq	$.L.str.0, 16(%rsp)
	movl	$.L.str.str, %edi
	movl	$.L.str.0, %esi
	xorl	%eax, %eax
	callq	printf@PLT
	movl	12(%rsp), %esi
	movl	$.L.str.int, %edi
	xorl	%eax, %eax
	callq	printf@PLT
	cmpl	$0, 12(%rsp)
	js	.LBB1_1
# %bb.2:                                # %if.end2
	addq	$24, %rsp
	.cfi_def_cfa_offset 8
	retq
.LBB1_1:                                # %if.then0
	.cfi_def_cfa_offset 32
	movl	$.L.str.str, %edi
	movl	$.L.str.1, %esi
	xorl	%eax, %eax
	callq	printf@PLT
	addq	$24, %rsp
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end1:
	.size	main, .Lfunc_end1-main
	.cfi_endproc
                                        # -- End function
	.type	.L.str.int,@object              # @.str.int
	.section	.rodata.str1.1,"aMS",@progbits,1
.L.str.int:
	.asciz	"%d\n"
	.size	.L.str.int, 4

	.type	.L.str.float,@object            # @.str.float
.L.str.float:
	.asciz	"%f\n"
	.size	.L.str.float, 4

	.type	.L.str.str,@object              # @.str.str
.L.str.str:
	.asciz	"%s\n"
	.size	.L.str.str, 4

	.type	.L.str.0,@object                # @.str.0
.L.str.0:
	.asciz	"teste"
	.size	.L.str.0, 6

	.type	.L.str.1,@object                # @.str.1
.L.str.1:
	.asciz	"negativo"
	.size	.L.str.1, 9

	.section	".note.GNU-stack","",@progbits
