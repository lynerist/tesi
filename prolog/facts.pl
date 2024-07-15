requisites(ToCheck) :-
    forall(requires(ToCheck, Thing), 
            exists(Thing)).

exists(Thing) :-
    provides(Provider, Thing), requisites(Provider).

valid(ToCheck) :-
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
provides(a->xb, aDef).
provides(b->y,  bDef).
provides(b->y,  bEnd).
provides(a->xa, aDef).
provides(b->ya, bDef).



