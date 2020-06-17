// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// The hacked top-most function.
// returns to goexit+PCQuantum.
TEXT ·hackedGoexit(SB),NOSPLIT,$0-0
	BYTE	$0x90	// NOP
	CALL	·hackedGoexit1(SB)	// does not return
	// traceback from hackedGoexit1 must hit code range of hackedGoexit.
	BYTE	$0x90	// NOP
