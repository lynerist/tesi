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

requisitesOne(ToCheck) :-
    requisitesOne(ToCheck, []).

requisitesOne(ToCheck, CantProvide) :-
    \+ (requiresOne(ToCheck, _, GroupID), 
        requiresOne(ToCheck, Thing, GroupID),
            exists([ToCheck | CantProvide], Thing),
        requiresOne(ToCheck, Another, GroupID),
            exists([ToCheck | CantProvide], Another),
            Thing \= Another
        ).

exists(Thing) :-
    exists([], Thing).

exists(CantProvide, Thing) :-
    provides(Provider, Thing), 
    \+ member(Provider, CantProvide),
    valid(Provider, CantProvide).

valid(ToCheck, CantProvide) :-
    provides(ToCheck, _),
    requisitesAll(ToCheck, CantProvide),
    requisitesNot(ToCheck, CantProvide),
    requisitesAny(ToCheck, CantProvide),
    requisitesOne(ToCheck, CantProvide).%,!.

valid(ToCheck) :-
    valid(ToCheck, []).

couldExist(Thing) :-
    provides(Provider, Thing).