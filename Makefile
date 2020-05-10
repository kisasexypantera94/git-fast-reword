build:
	go build

test: build git-reset-testdata
	go test ./...

git-reset-testdata:
	cd utilite/testdata/django-like-queryset ; git reset --hard