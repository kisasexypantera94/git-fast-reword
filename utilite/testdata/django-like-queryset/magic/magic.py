from __future__ import annotations

import operator
from functools import reduce
from typing import *


class Magic:
    __relations = {
        "eq": lambda y: lambda x: x == y,
        "lt": lambda y: lambda x: x < y,
        "lte": lambda y: lambda x: x <= y,
        "gte": lambda y: lambda x: x >= y,
        "gt": lambda y: lambda x: x > y,
        "ne": lambda y: lambda x: x != y,
        "in": lambda y: lambda x: x in y,
        "startswith": lambda y: lambda x: x.startswith(y)
    }

    def __init__(self: Magic, data: Iterable, predicate: Callable[bool] = lambda x: True) -> None:
        self.data = data
        self.predicate = predicate

    def __iter__(self: Magic):
        for x in self.data:
            try:
                if self.predicate(x):
                    yield x
            except TypeError:
                pass

    def filter_(self: Magic, *args: Magic, **kwargs: object) -> Magic:
        predicate_from_query = self.__process_query(args, kwargs)
        return self.__new_magic(operator.and_, predicate_from_query)

    def or_(self: Magic, *args: Magic, **kwargs: object) -> Magic:
        predicate_from_query = self.__process_query(args, kwargs)
        return self.__new_magic(operator.or_, predicate_from_query)

    def not_(self: Magic, *args: Magic, **kwargs: object) -> Magic:
        predicate_from_query = self.__process_query(args, kwargs)
        return self.__new_magic(operator.and_, lambda x: not predicate_from_query(x))

    def __process_query(self: Magic, predicates: Tuple[Magic], conditions: Dict[str, object]) -> Callable[bool]:
        query = self.__parse_query(conditions) + [m.predicate for m in predicates]
        folded_query = self.__fold_query(query)
        return folded_query

    def __parse_query(self: Magic, query: Dict[str, object]) -> List[Callable[bool]]:
        return [self.__make_function(name, y) for name, y in query.items()]

    def __make_function(self: Magic, name: str, y: object) -> Callable[bool]:
        """Парсинг условия и преобразование его в функцию"""
        attributes = name.split("__")
        relation_with_y = self.__get_relation(attributes[-1], y)

        def f(x: object) -> bool:
            cur = x
            for attr in attributes[:-1]:
                if not attr:
                    break

                if hasattr(cur, attr):
                    cur = getattr(cur, attr)
                else:
                    return False

            return relation_with_y(cur)

        return f

    def __get_relation(self: Magic, name: str, y: object) -> Callable[bool]:
        return self.__relations[name](y)

    @staticmethod
    def __fold_query(query: List[Callable[bool]]) -> Callable[bool]:
        """Логическое перемножение булевых функций"""

        def folded_query(x: object):
            if len(query) == 1:
                return query[0](x)

            return reduce(lambda prev, cur: prev(x) and cur(x), query)

        return folded_query

    def __new_magic(self: Magic, op: Callable[bool], new: Callable[bool]) -> Magic:
        """Создание обновленного объекта `Magic`"""

        def new_predicate(x: object, old: Callable[bool] = self.predicate):
            return op(old(x), new(x))

        return Magic(self.data, new_predicate)
