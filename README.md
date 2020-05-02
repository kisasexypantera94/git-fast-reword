# git-fast-reword

## Dependencies
You should install libgit2 before building  

#### macOS
```zsh
brew install libgit2
```

## Building 
```zsh
$ go build
```

## Usage
Run in directory with .git:
```zsh
$  ./git-fast-reword -h
NAME:
   git-fast-reword - git-fast-reword hash new_message

USAGE:
   git-fast-reword [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)

$ ./git-fast-reword HEAD~123 "fyi"
```