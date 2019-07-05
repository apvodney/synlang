id=mod=ramp
	freq=20
	out=sin.freq

% id=modname=sin inputs=[freq] outputs=[out]
	freq=ramp.out
	out=.out
% {
	id=mod=ramp
		freq=(440 + .freq * 1000)
		out=sinshp.in
	id=mod=sinshp
		in=ramp.out
		out=.out
}
