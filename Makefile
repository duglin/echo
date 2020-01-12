all: echo .echo

echo: echo.go
	go build -o echo echo.go

.echo: echo.go
	go build -o /dev/null echo.go # quick fail
	docker build -t duglin/echo .
	docker push duglin/echo
	touch .echo

clean:
	rm -f .echo echo
