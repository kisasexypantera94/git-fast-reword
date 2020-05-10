# git-fast-reword

## Зависимости
* libgit2  

#### macOS
```zsh
brew install libgit2
```

## Сборка и установка 
```zsh
$ go install
```

## Запуск и тестирование
```zsh
➜  git-fast-reword git:(master) ✗ make test
go build
cd utilite/testdata/django-like-queryset ; git reset --hard
HEAD сейчас на be51798 Update README.md
go test ./...
?       git-fast-reword [no test files]
ok      git-fast-reword/utilite 0.454s

➜  git-fast-reword git:(master) ✗ cd intellij-community 
➜  intellij-community git:(master) git-fast-reword -h
NAME:
   git-fast-reword - git-fast-reword hash new_message

USAGE:
   git-fast-reword [global options] command [command options] [arguments...]

COMMANDS:
   from-file, ff  
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
 
➜  intellij-community git:(master) cat test.json 
{
  "HEAD~31": "renamed HEAD~31",
  "HEAD~53": "renamed HEAD~53",
  "HEAD~173": "renamed HEAD~173"
}

➜  intellij-community git:(master) ✗ time git-fast-reword ff test.json
New hashes:
{
  "HEAD~173": "5d4cf25f2a99cd0fda742e206b109a9d19b2fe54",
  "HEAD~31": "24970821f817f3e2a36e9f6215b3a4e037244e0b",
  "HEAD~53": "f8eac669e619493ad7026d64151a127c90102d44"
}
git-fast-reword ff test.json  0,02s user 0,02s system 17% cpu 0,201 total

➜  intellij-community git:(master) git cat-file -p HEAD~31
tree 9876e6a8a3d9eb076d5787b4e54478409662ad7d
parent 77a3848975d4cbf80f54e356570c2a20c9f44683
author Yuriy Artamonov <yuriy.artamonov@jetbrains.com> 1588258011 +0300
committer intellij-monorepo-bot <intellij-monorepo-bot-no-reply@jetbrains.com> 1588326021 +0000

renamed HEAD~31

➜  intellij-community git:(master) git cat-file -p HEAD~53
tree c6d98eaac03c46a1824b578876165f68ae155aa5
parent e24ba542c515b53320853e5ba4f724e790442186
author Vladislav.Soroka <Vladislav.Soroka@jetbrains.com> 1588258289 +0300
committer intellij-monorepo-bot <intellij-monorepo-bot-no-reply@jetbrains.com> 1588326021 +0000

renamed HEAD~53

➜  intellij-community git:(master) git cat-file -p HEAD~173
tree 7a035ec1c133f15d8850da4362324d93f38354d5
parent 222aad0c9aa95fb671000476c035f384a1de39a2
author Semyon Proshev <Semyon.Proshev@jetbrains.com> 1587481536 +0300
committer intellij-monorepo-bot <intellij-monorepo-bot-no-reply@jetbrains.com> 1588197868 +0000

renamed HEAD~173

# случай, описанный ниже
➜  intellij-community git:(master) ✗ time git-fast-reword 13b78e06c18e2da98674b688e56df0b53b9fed76 "s bogom"
New hashes:
{
  "13b78e06c18e2da98674b688e56df0b53b9fed76": "8abaf5f64dee96bf5f1256daff3b3f6d32b1c070"
}
git-fast-reword 13b78e06c18e2da98674b688e56df0b53b9fed76 "s bogom"  6,52s user 0,73s system 92% cpu 7,818 total

➜  intellij-community git:(master) ✗ git cat-file -p 8abaf5f64dee96bf5f1256daff3b3f6d32b1c070
tree 60210d988e0eeb730544264e77a20b7cdda3ceb5
parent 28adfdb1e958083068562aa9909437d4db51f312
author Ivan Chirkov <Ivan.Chirkov@jetbrains.com> 1587490014 +0200
committer intellij-monorepo-bot <intellij-monorepo-bot-no-reply@jetbrains.com> 1587540666 +0000

s bogom

```

## Алгоритм
Утилита поиском в глубину обходит коммиты и обновляет их по надобности.
Рекурсия прекращается как только посещены все коммиты с новыми сообщениями,
поэтому в большинстве случаев утилита обойдет лишь малую часть графа.
Поскольку возможно существование пути между любыми двумя коммитами, то также
возможна такая ситуация:  
![](assets/bad_case.png)  
Допустим мы хотим переименовать коммит `2`. 
Поиск будет идти в следующем порядке: `0->1->2` – тут мы увидели,
что посетили все коммиты с новыми сообщениями (в этой ветке) и прерываем рекурсию.
Далее поиск пойдет в вершину `3` и так как мы не знаем наперед есть ли путь из этой
вершины в вершину `2`, то нам придется обойти весь граф до самого дна.
Ориентироваться на время коммита (как commiter, так и author) не вариант, поскольку
оно могло быть изменено до этого командой `rebase`. Я так и не придумал как можно
гарантировано избежать спуска на дно.
Поскольку даже в худшем случае утилита работает в пределах 10-13 секунд на репозитории
`IntelliJ IDEA Community Edition`, то я не стал париться и оставил все как есть.