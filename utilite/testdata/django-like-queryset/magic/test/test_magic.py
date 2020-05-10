from unittest import TestCase

from magic.magic import Magic


class Framework:
    def __init__(self, framework, language, type_):
        self.framework = framework
        self.language = language
        self.type = type_


class TestMagic(TestCase):
    def test_integration(self):
        a = Magic(range(100))
        a = a.filter_(__gt=3).filter_(__lte=5)
        self.assertEqual([4, 5], list(a))

        a = a.or_(__lt=20, __gt=13).not_(__eq=15)
        self.assertEqual([4, 5, 14, 16, 17, 18, 19], list(a))

        a = a.filter_(__gt=3).not_(__gte=4, __lte=5)
        self.assertEqual([14, 16, 17, 18, 19], list(a))

        a2 = Magic(range(50))
        a2 = a2.not_(__gte=16, __lte=18)

        a = a.filter_(a2).or_(__eq=99)
        self.assertEqual([14, 19, 99], list(a))

        partial = Magic(range(100)).filter_(__gt=3, __lt=7)
        b = Magic(range(1, 100)).filter_(__lte=10).not_(partial)
        self.assertEqual([1, 2, 3, 7, 8, 9, 10], list(b))

    def test_multiple_attributes(self):
        data = [Framework('Django', 'Python', 'full-stack'),
                "У этого объекта нет аттрибута `framework`",
                Framework('Rails', 'Ruby', 'full-stack'),
                Framework('Sinatra', 'Ruby', 'micro'),
                Framework('Zend', 'PHP', 'full-stack'),
                Framework('Slim', 'PHP', 'micro')]

        a = Magic(data)
        a = a.filter_(framework__startswith='S')
        self.assertEqual([data[3], data[5]], list(a))
