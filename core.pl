requisitesAll(ToCheck) :-
    requisitesAll(ToCheck, []).

requisitesAll(ToCheck, CantProvide) :-
	findall(Thing, requiresAll(ToCheck, Thing), NeededThings), 
	maplist(exists([ToCheck | CantProvide]), NeededThings),!.

exists(Thing) :-
    exists([], Thing).

exists(CantProvide, Thing) :-
    provides(Provider, Thing), 
    \+ member(Provider, CantProvide),
    valid(Provider, CantProvide).

valid(ToCheck, CantProvide) :-
    provides(ToCheck, _),
    requisitesAll(ToCheck, CantProvide),!.

valid(ToCheck) :-
    provides(ToCheck, _), 
    requisitesAll(ToCheck),!.