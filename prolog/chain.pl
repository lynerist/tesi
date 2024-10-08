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
requires(a->b,  bDef).
requires(a->b,  bEnd).
requires(b->c,  cDef).
requires(b->c,  cEnd).

provides(p->a,  pDef).
provides(a->b,  aDef).
provides(b->c,  bDef).
provides(c->x,  cDef).
provides(c->x,  cEnd).

provides(b->c, bEnd).
provides(a->b, aEnd).
