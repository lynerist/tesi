requisitesAll(ToCheck) :-
    requisitesAll(ToCheck, []).

requisitesAll(ToCheck, CantProvide) :-
    \+ (requiresAll(ToCheck, Thing), 
        \+ exists([ToCheck | CantProvide], Thing)).

requisitesNot(ToCheck) :-
    requisitesNot(ToCheck, []).

requisitesNot(ToCheck, CantProvide) :-
    \+ (requiresNot(ToCheck, Thing), exists([ToCheck | CantProvide], Thing)).

requisitesAny(ToCheck) :-
    requisitesAny(ToCheck, []).

requisitesAny(ToCheck, CantProvide) :-
    \+ (requiresAny(ToCheck, _, GroupID), 
        \+ (requiresAny(ToCheck, Thing, GroupID),
            exists([ToCheck | CantProvide], Thing))).

exists(Thing) :-
    exists([], Thing).

exists(CantProvide, Thing) :-
    provides(Provider, Thing), 
    \+ member(Provider, CantProvide),
    valid(Provider, CantProvide),!.

valid(ToCheck, CantProvide) :-
    provides(ToCheck, _),
    requisitesAll(ToCheck, CantProvide),
    requisitesNot(ToCheck, CantProvide),
    requisitesAny(ToCheck, CantProvide),!.

valid(ToCheck) :-
    valid(ToCheck, []).