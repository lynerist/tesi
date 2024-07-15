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
requires(c->a,  aDef).
requires(c->a,  aEnd).

%requires(b->d,  dDef).
%requires(b->d,  dEnd).

provides(p->a,  pDef).
provides(p->a,  pEnd).
provides(a->b,  aDef).
provides(a->b,  aEnd).
provides(b->c,  bDef).
provides(b->c,  bEnd).
provides(c->a,  cDef).
provides(c->a,  cEnd).

%provides(d->x, dDef).
%provides(d->x, dEnd).
%provides(b->d, bDef).
%provides(b->d, bEnd).
