# django-like-queryset

## About
QuerySet implementation in django like style. The `Magic` object accumulates a predicate in the form of a function that is called only at the stage of iteration.

## Running
```
$ python3 -m unittest magic/test/test_magic.py

$ python3
>>> from magic.magic import Magic
>>> a = Magic(range(100))
>>> a_filtered = a.filter_(__gt=3).filter_(__lte=5).or_(__lt=20, __gt=17).not_(__eq=15)
>>> list(a_filtered)
[4, 5, 18, 19]
```
