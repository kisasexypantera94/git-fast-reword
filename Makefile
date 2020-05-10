build:
	go build

test: build git-clone-testdata git-reset-testdata
	go test ./...

git-clone-testdata:
	if ! cd utilite/testdata/django-like-queryset; then\
		git clone https://github.com/kisasexypantera94/django-like-queryset.git utilite/testdata/django-like-queryset;\
	fi

git-reset-testdata:
	cd utilite/testdata/django-like-queryset ; git reset --hard origin/master
