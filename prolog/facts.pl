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
    provides(ToCheck, _), %può non avere provides una feature?
    requisites(ToCheck).

requires(p->a,  aDef).
requires(p->a,  aEnd).
requires(a->xb, bDef).
requires(a->xb, bEnd).
requires(a->xa, aDef).
requires(a->xa, aEnd).
requires(b->ya, aDef).
requires(b->ya, aEnd).

provides(p->a,  pDef).
provides(p->a,  pEnd).
provides(a->xb, aDef).
provides(a->xb, aEnd).
provides(b->y,  bDef).
provides(b->y,  bEnd).
provides(a->xa, aDef).
provides(b->ya, bDef).



