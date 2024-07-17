
requisites(ToCheck) :-
        requisites(ToCheck, []).

requisites(ToCheck, CantProvide) :-
        forall(requires(ToCheck, Thing), 
                        exists(Thing, [ToCheck | CantProvide])).

exists(Thing) :-
        exists(Thing, []).

exists(Thing, CantProvide) :-
        provides(Provider, Thing), 
        \+ member(Provider, CantProvide),
        requisites(Provider, CantProvide).

valid(ToCheck) :-
        provides(ToCheck, _), 
        requisites(ToCheck).


requires('812598aedbb255944247d7d89d1063d6','df9eb5119068f09cb467176dd67830d7').
requires('812598aedbb255944247d7d89d1063d6','ca3b88403be6b0de57b6182e744d16c0').
requires('9ea5d25c0121034f832ef10059b44c09','5c3c6bd8ea22dece2d8749aa20abbd1a').
requires('9ea5d25c0121034f832ef10059b44c09','4f88a04c300fb1732cdc6b6b0a67b60d').
requires('eb769bf826cf66801dce950c4bc30ab4','df9eb5119068f09cb467176dd67830d7').
requires('eb769bf826cf66801dce950c4bc30ab4','ca3b88403be6b0de57b6182e744d16c0').
requires('a1477979a0b323749d3aa6a415905041','df9eb5119068f09cb467176dd67830d7').
requires('a1477979a0b323749d3aa6a415905041','ca3b88403be6b0de57b6182e744d16c0').

provides('812598aedbb255944247d7d89d1063d6','820b45c94255f88a0f8ce1c62a354b63').
provides('812598aedbb255944247d7d89d1063d6','f407bc992a3949f356be8fb55c6bf227').
provides('9ea5d25c0121034f832ef10059b44c09','df9eb5119068f09cb467176dd67830d7').
provides('9ea5d25c0121034f832ef10059b44c09','ca3b88403be6b0de57b6182e744d16c0').
provides('914c338f8ecb57d7794b1279d9e5e138','5c3c6bd8ea22dece2d8749aa20abbd1a').
provides('914c338f8ecb57d7794b1279d9e5e138','4f88a04c300fb1732cdc6b6b0a67b60d').
provides('eb769bf826cf66801dce950c4bc30ab4','df9eb5119068f09cb467176dd67830d7').
provides('a1477979a0b323749d3aa6a415905041','5c3c6bd8ea22dece2d8749aa20abbd1a').
provides('a1477979a0b323749d3aa6a415905041','4f88a04c300fb1732cdc6b6b0a67b60d').