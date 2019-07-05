id=mod1 mod=ramp freq=20 out=sin.freq

{ modname=sin inputs=[freq] outputs=[out] id=osc1 out=.out

	mod=ramp freq=(440 + .freq * 1000) out=sinshp.in
	
	mod=sinshp in=ramp.out out=.out
}
