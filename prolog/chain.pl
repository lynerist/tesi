requisites(ToCheck) :-
    forall(requires(ToCheck, Thing), 
            exists(Thing)).

exists(Thing) :-
    provides(Provider, Thing), requisites(Provider).

valid(ToCheck) :-
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
