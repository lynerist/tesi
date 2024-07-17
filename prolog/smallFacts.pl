valid(X) :-
    findall(Y, needs(X, Y), Needed),
    maplist(exists, Needed).

exists(a).
exists(b).

needs(x,a).
needs(x,b).
needs(y,b).
needs(y,c).